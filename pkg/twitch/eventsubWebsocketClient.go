package twitch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"strings"
	"time"

	"github.com/Luzifer/go_helpers/backoff"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
)

const (
	eventsubLiveSocketDest       = "wss://eventsub.wss.twitch.tv/ws"
	socketConnectTimeout         = 15 * time.Second
	socketInitialTimeout         = 30 * time.Second
	socketTimeoutGraceMultiplier = 1.5

	retrySubscribeMaxTotal = 30 * time.Minute
	retrySubscribeMaxWait  = 5 * time.Minute
	retrySubscribeMinWait  = 30 * time.Second
)

const (
	eventsubSocketMessageTypeKeepalive    = "session_keepalive"
	eventsubSocketMessageTypeNotification = "notification"
	eventsubSocketMessageTypeReconnect    = "session_reconnect"
	eventsubSocketMessageTypeWelcome      = "session_welcome"
)

const (
	eventsubCloseCodeInternalServerError  = 4000
	eventsubCloseCodeClientSentTraffic    = 4001
	eventsubCloseCodeClientFailedPingPong = 4002
	eventsubCloseCodeConnectionUnused     = 4003
	eventsubCloseCodeReconnectGraceExpire = 4004
	eventsubCloseCodeNetworkTimeout       = 4005
	eventsubCloseCodeNetworkError         = 4006
	eventsubCloseCodeInvalidReconnect     = 4007
)

type (
	// EventSubSocketClient manages a WebSocket transport for the Twitch
	// EventSub API
	EventSubSocketClient struct {
		logger            *logrus.Entry
		socketDest        string
		socketID          string
		subscriptionTypes []eventSubSocketSubscriptionType
		twitch            *Client

		conn    *websocket.Conn
		newconn *websocket.Conn

		runCtx       context.Context //nolint:containedctx // internally held context for this client
		runCtxCancel context.CancelFunc
	}

	// EventSubSocketClientOpt is a setter function to apply changes to
	// the EventSubSocketClient on create
	EventSubSocketClientOpt func(*EventSubSocketClient)

	eventSubSocketMessage struct {
		Metadata struct {
			MessageID           string    `json:"message_id"`
			MessageType         string    `json:"message_type"`
			MessageTimestamp    time.Time `json:"message_timestamp"`
			SubscriptionType    string    `json:"subscription_type"`
			SubscriptionVersion string    `json:"subscription_version"`
		} `json:"metadata"`
		Payload json.RawMessage `json:"payload"`
	}

	eventSubSocketSubscriptionType struct {
		Event, Version  string
		Condition       EventSubCondition
		Callback        func(json.RawMessage) error
		BackgroundRetry bool
	}

	eventSubSocketPayloadNotification struct {
		Event        json.RawMessage `json:"event"`
		Subscription struct {
			ID        string            `json:"id"`
			Status    string            `json:"status"`
			Type      string            `json:"type"`
			Version   string            `json:"version"`
			Cost      int64             `json:"cost"`
			Condition EventSubCondition `json:"condition"`
			Transport struct {
				Method    string `json:"method"`
				SessionID string `json:"session_id"`
			} `json:"transport"`
			CreatedAt time.Time `json:"created_at"`
		} `json:"subscription"`
	}

	eventSubSocketPayloadSession struct {
		Session struct {
			ID                      string    `json:"id"`
			Status                  string    `json:"status"`
			ConnectedAt             time.Time `json:"connected_at"`
			KeepaliveTimeoutSeconds int64     `json:"keepalive_timeout_seconds"`
			ReconnectURL            *string   `json:"reconnect_url"`
		} `json:"session"`
	}
)

// NewEventSubSocketClient creates a new EventSubSocketClient and
// applies the given EventSubSocketClientOpts
func NewEventSubSocketClient(opts ...EventSubSocketClientOpt) (*EventSubSocketClient, error) {
	ctx, cancel := context.WithCancel(context.Background())

	c := &EventSubSocketClient{
		runCtx:       ctx,
		runCtxCancel: cancel,
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.socketDest == "" {
		c.socketDest = eventsubLiveSocketDest
	}

	if c.logger == nil {
		discardLogger := logrus.New()
		discardLogger.SetOutput(io.Discard)
		c.logger = logrus.NewEntry(discardLogger)
	}

	if c.twitch == nil {
		return nil, errors.New("no twitch-client configured")
	}

	return c, nil
}

// WithLogger configures the logger within the EventSubSocketClient
func WithLogger(logger *logrus.Entry) EventSubSocketClientOpt {
	return func(e *EventSubSocketClient) { e.logger = logger }
}

// WithMustSubscribe adds a topic to the subscriptions to be done on
// connect
func WithMustSubscribe(event, version string, condition EventSubCondition, callback func(json.RawMessage) error) EventSubSocketClientOpt {
	if version == "" {
		version = EventSubTopicVersion1
	}

	return func(e *EventSubSocketClient) {
		e.subscriptionTypes = append(e.subscriptionTypes, eventSubSocketSubscriptionType{
			Event:     event,
			Version:   version,
			Condition: condition,
			Callback:  callback,
		})
	}
}

// WithRetryBackgroundSubscribe adds a topic to the subscriptions to
// be done on connect async
func WithRetryBackgroundSubscribe(event, version string, condition EventSubCondition, callback func(json.RawMessage) error) EventSubSocketClientOpt {
	if version == "" {
		version = EventSubTopicVersion1
	}

	return func(e *EventSubSocketClient) {
		e.subscriptionTypes = append(e.subscriptionTypes, eventSubSocketSubscriptionType{
			Event:           event,
			Version:         version,
			Condition:       condition,
			Callback:        callback,
			BackgroundRetry: true,
		})
	}
}

// WithSocketURL overwrites the socket URL to connect to
func WithSocketURL(url string) EventSubSocketClientOpt {
	return func(e *EventSubSocketClient) { e.socketDest = url }
}

// WithTwitchClient overwrites the Client to be used
func WithTwitchClient(c *Client) EventSubSocketClientOpt {
	return func(e *EventSubSocketClient) { e.twitch = c }
}

// Close cancels the contained context and brings the
// EventSubSocketClient to a halt
func (e *EventSubSocketClient) Close() { e.runCtxCancel() }

// Run starts the main communcation loop for the EventSubSocketClient
//
//nolint:gocyclo // Makes no sense to split further
func (e *EventSubSocketClient) Run() error {
	var (
		errC             = make(chan error, 1)
		keepaliveTimeout = socketInitialTimeout
		msgC             = make(chan eventSubSocketMessage, 1)
		timeoutC         = make(chan struct{}, 1)
		socketTimeout    = newKeepaliveTracker(timeoutC, keepaliveTimeout)
	)

	if err := e.connect(e.socketDest, msgC, errC, "client init"); err != nil {
		return fmt.Errorf("establishing initial connection: %w", err)
	}

	defer func() {
		if err := e.conn.Close(); err != nil {
			e.logger.WithError(helpers.CleanNetworkAddressFromError(err)).Error("finally closing socket")
		}
	}()

	for {
		select {
		case err := <-errC:
			// Something went wrong
			if err = e.handleSocketError(err, msgC, errC); err != nil {
				return err
			}

		case <-timeoutC:
			// No message received, deeming connection dead
			e.logger.WithFields(logrus.Fields{
				"expired":    socketTimeout.ExpiresAt(),
				"last_event": socketTimeout.LastRenew(),
			}).Warn("eventsub socket missed keepalive")

			socketTimeout = newKeepaliveTracker(timeoutC, socketInitialTimeout)
			if err := e.connect(e.socketDest, msgC, errC, "socket timeout"); err != nil {
				errC <- fmt.Errorf("re-connecting after timeout: %w", err)
				continue
			}

		case msg := <-msgC:
			// The keepalive timer is reset with each notification or
			// keepalive message.
			socketTimeout.Renew(keepaliveTimeout)

			switch msg.Metadata.MessageType {
			case eventsubSocketMessageTypeKeepalive:
				// Handle only for debug, timer reset is done above
				e.logger.Trace("keepalive received")

			case eventsubSocketMessageTypeNotification:
				// We got mail! Yay!
				if err := e.handleNotificationMessage(msg); err != nil {
					errC <- err
				}

			case eventsubSocketMessageTypeReconnect:
				// Twitch politely asked us to reconnect
				if err := e.handleReconnectMessage(msg, msgC, errC); err != nil {
					errC <- err
				}

			case eventsubSocketMessageTypeWelcome:
				var err error
				if keepaliveTimeout, err = e.handleWelcomeMessage(msg); err != nil {
					errC <- err
				}

			default:
				e.logger.WithField("type", msg.Metadata.MessageType).Error("unknown message type received")
			}

		case <-e.runCtx.Done():
			return nil
		}
	}
}

func (e *EventSubSocketClient) connect(url string, msgC chan eventSubSocketMessage, errC chan error, reason string) error {
	e.logger.WithField("reason", reason).Debug("(re-)connecting websocket")

	ctx, cancel := context.WithTimeout(context.Background(), socketConnectTimeout)
	defer cancel()

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil) //nolint:bodyclose // Close is implemented at other place
	if err != nil {
		return fmt.Errorf("dialing websocket: %w", err)
	}

	go func() {
		for {
			var msg eventSubSocketMessage
			if err = conn.ReadJSON(&msg); err != nil {
				errC <- fmt.Errorf("reading message: %w", err)
				return
			}

			msgC <- msg
		}
	}()

	e.newconn = conn
	return nil
}

func (e *EventSubSocketClient) handleNotificationMessage(msg eventSubSocketMessage) error {
	var payload eventSubSocketPayloadNotification
	if err := msg.Unmarshal(&payload); err != nil {
		return fmt.Errorf("unmarshalling notification: %w", err)
	}

	for _, st := range e.subscriptionTypes {
		if st.Event != payload.Subscription.Type || st.Version != payload.Subscription.Version || !reflect.DeepEqual(st.Condition, payload.Subscription.Condition) {
			continue
		}

		if err := st.Callback(payload.Event); err != nil {
			e.logger.WithError(err).WithFields(logrus.Fields{
				"condition": st.Condition,
				"event":     st.Event,
				"version":   st.Version,
			}).Error("callback caused error")
		}
	}

	return nil
}

func (e *EventSubSocketClient) handleReconnectMessage(msg eventSubSocketMessage, msgC chan eventSubSocketMessage, errC chan error) error {
	e.logger.Debug("socket ask for reconnect")

	var payload eventSubSocketPayloadSession
	if err := msg.Unmarshal(&payload); err != nil {
		return fmt.Errorf("unmarshalling reconnect message: %w", err)
	}

	if payload.Session.ReconnectURL == nil {
		return errors.New("reconnect message did not contain reconnect_url")
	}

	if err := e.connect(*payload.Session.ReconnectURL, msgC, errC, "reconnect requested"); err != nil {
		return fmt.Errorf("re-connecting after reconnect message: %w", err)
	}

	return nil
}

//nolint:gocyclo // just reacting on websocket events
func (e *EventSubSocketClient) handleSocketError(err error, msgC chan eventSubSocketMessage, errC chan error) error {
	var closeErr *websocket.CloseError
	if errors.As(err, &closeErr) {
		switch closeErr.Code {
		case eventsubCloseCodeInternalServerError:
			e.logger.Warn("websocket reported internal server error")
			if err = e.connect(e.socketDest, msgC, errC, "internal-server-error"); err != nil {
				return fmt.Errorf("re-connecting after internal-server-error: %w", err)
			}

			return nil

		case eventsubCloseCodeClientSentTraffic:
			e.logger.Error("wrong usage of websocket (client-sent-traffic)")

		case eventsubCloseCodeClientFailedPingPong:
			e.logger.Error("wrong usage of websocket (missing-ping-pong)")

		case eventsubCloseCodeConnectionUnused:
			e.logger.Error("wrong usage of websocket (no-topics-subscribed)")

		case eventsubCloseCodeReconnectGraceExpire:
			e.logger.Error("wrong usage of websocket (no-reconnect-in-time)")

		case eventsubCloseCodeNetworkTimeout:
			e.logger.Warn("websocket reported network timeout")
			if err = e.connect(e.socketDest, msgC, errC, "network-timeout"); err != nil {
				return fmt.Errorf("re-connecting after network-timeout: %w", err)
			}

			return nil

		case eventsubCloseCodeNetworkError:
			e.logger.Warn("websocket reported network error")
			if err = e.connect(e.socketDest, msgC, errC, "network-error"); err != nil {
				return fmt.Errorf("re-connecting after network-error: %w", err)
			}

			return nil

		case eventsubCloseCodeInvalidReconnect:
			e.logger.Warn("websocket reported invalid reconnect url")

		case websocket.CloseNormalClosure:
			// We don't take action here as a graceful close should
			// be initiated by us after establishing a new conn
			e.logger.Debug("websocket was closed normally")
			return nil

		case websocket.CloseAbnormalClosure:
			e.logger.Warn("websocket reported abnormal closure")
			if err = e.connect(e.socketDest, msgC, errC, "network-error"); err != nil {
				return fmt.Errorf("re-connecting after abnormal closure: %w", err)
			}

			return nil

		default:
			// Some non-twitch close code we did not expect
			e.logger.WithError(closeErr).Error("websocket reported unexpected error code")
		}
	}

	if errors.Is(err, net.ErrClosed) {
		// This isn't nice but might happen, in this  case the socket is
		// already gone but the read didn't notice that until this error
		return nil
	}

	return err
}

func (e *EventSubSocketClient) handleWelcomeMessage(msg eventSubSocketMessage) (time.Duration, error) {
	var payload eventSubSocketPayloadSession
	if err := msg.Unmarshal(&payload); err != nil {
		return socketInitialTimeout, fmt.Errorf("unmarshalling welcome message: %w", err)
	}

	// Close old connection if present
	if e.conn != nil {
		if err := e.conn.Close(); err != nil {
			e.logger.WithError(err).Error("closing old websocket")
		}
	}

	// Promote new connection to existing conn
	e.conn, e.newconn = e.newconn, nil

	// Subscribe to topics if the socket ID changed (should only
	// happen on first connect or if we established a new
	// connection after something broke)
	if e.socketID != payload.Session.ID {
		e.socketID = payload.Session.ID
		if err := e.subscribeAll(); err != nil {
			return socketInitialTimeout, fmt.Errorf("subscribing to topics: %w", err)
		}
	}

	e.logger.WithField("id", e.socketID).Debug("websocket connected successfully")

	// Configure proper keepalive
	return time.Duration(float64(payload.Session.KeepaliveTimeoutSeconds)*socketTimeoutGraceMultiplier) * time.Second, nil
}

func (e *EventSubSocketClient) retryBackgroundSubscribe(st eventSubSocketSubscriptionType) {
	err := backoff.NewBackoff().
		WithMaxIterationTime(retrySubscribeMaxWait).
		WithMaxTotalTime(retrySubscribeMaxTotal).
		WithMinIterationTime(retrySubscribeMinWait).
		Retry(func() error {
			if err := e.runCtx.Err(); err != nil {
				// Our run-context was cancelled, stop retrying to subscribe
				// to topics as this client was closed
				return backoff.NewErrCannotRetry(err)
			}

			return e.subscribe(st)
		})
	if err != nil {
		e.logger.
			WithError(err).
			WithField("topic", strings.Join([]string{st.Event, st.Version}, "/")).
			Error("gave up retrying to subscribe")
	}
}

func (e *EventSubSocketClient) subscribe(st eventSubSocketSubscriptionType) error {
	logger := e.logger.
		WithField("topic", strings.Join([]string{st.Event, st.Version}, "/"))

	if _, err := e.twitch.createEventSubSubscriptionWebsocket(context.Background(), eventSubSubscription{
		Type:      st.Event,
		Version:   st.Version,
		Condition: st.Condition,
		Transport: eventSubTransport{
			Method:    "websocket",
			SessionID: e.socketID,
		},
	}); err != nil {
		logger.WithError(err).Debug("subscribing to topic")
		return fmt.Errorf("subscribing to %s/%s: %w", st.Event, st.Version, err)
	}

	logger.
		WithField("topic", strings.Join([]string{st.Event, st.Version}, "/")).
		Debug("subscribed to topic")
	return nil
}

func (e *EventSubSocketClient) subscribeAll() (err error) {
	for i := range e.subscriptionTypes {
		st := e.subscriptionTypes[i]

		if st.BackgroundRetry {
			go e.retryBackgroundSubscribe(st)
			continue
		}

		if err = e.subscribe(st); err != nil {
			return err
		}
	}

	return nil
}

func (e eventSubSocketMessage) Unmarshal(dest any) error {
	if err := json.Unmarshal(e.Payload, dest); err != nil {
		return fmt.Errorf("unmarshalling payload: %w", err)
	}

	return nil
}

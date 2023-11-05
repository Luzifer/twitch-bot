package twitch

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"reflect"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
)

const (
	eventsubLiveSocketDest       = "wss://eventsub.wss.twitch.tv/ws"
	socketConnectTimeout         = 15 * time.Second
	socketInitialTimeout         = 30 * time.Second
	socketTimeoutGraceMultiplier = 1.5
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
	EventSubSocketClient struct {
		logger            *logrus.Entry
		socketDest        string
		socketID          string
		subscriptionTypes []eventSubSocketSubscriptionType
		twitch            *Client

		conn    *websocket.Conn
		newconn *websocket.Conn

		runCtx       context.Context
		runCtxCancel context.CancelFunc
	}

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
		Event, Version string
		Condition      EventSubCondition
		Callback       func(json.RawMessage) error
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

func WithLogger(logger *logrus.Entry) EventSubSocketClientOpt {
	return func(e *EventSubSocketClient) { e.logger = logger }
}

func WithSocketURL(url string) EventSubSocketClientOpt {
	return func(e *EventSubSocketClient) { e.socketDest = url }
}

func WithSubscription(event, version string, condition EventSubCondition, callback func(json.RawMessage) error) EventSubSocketClientOpt {
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

func WithTwitchClient(c *Client) EventSubSocketClientOpt {
	return func(e *EventSubSocketClient) { e.twitch = c }
}

func (e *EventSubSocketClient) Close() { e.runCtxCancel() }

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
		return errors.Wrap(err, "establishing initial connection")
	}

	defer func() {
		if err := e.conn.Close(); err != nil {
			e.logger.WithError(helpers.CleanOpError(err)).Error("finally closing socket")
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
				errC <- errors.Wrap(err, "re-connecting after timeout")
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
		return errors.Wrap(err, "dialing websocket")
	}

	go func() {
		for {
			var msg eventSubSocketMessage
			if err = conn.ReadJSON(&msg); err != nil {
				errC <- errors.Wrap(err, "reading message")
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
		return errors.Wrap(err, "unmarshalling notification")
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
		return errors.Wrap(err, "unmarshalling reconnect message")
	}

	if payload.Session.ReconnectURL == nil {
		return errors.New("reconnect message did not contain reconnect_url")
	}

	if err := e.connect(*payload.Session.ReconnectURL, msgC, errC, "reconnect requested"); err != nil {
		return errors.Wrap(err, "re-connecting after reconnect message")
	}

	return nil
}

func (e *EventSubSocketClient) handleSocketError(err error, msgC chan eventSubSocketMessage, errC chan error) error {
	var closeErr *websocket.CloseError
	if errors.As(err, &closeErr) {
		switch closeErr.Code {
		case eventsubCloseCodeInternalServerError:
			e.logger.Warn("websocket reported internal server error")
			return errors.Wrap(e.connect(e.socketDest, msgC, errC, "internal-server-error"), "re-connecting after internal-server-error")

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
			return errors.Wrap(e.connect(e.socketDest, msgC, errC, "network-timeout"), "re-connecting after network-timeout")

		case eventsubCloseCodeNetworkError:
			e.logger.Warn("websocket reported network error")
			return errors.Wrap(e.connect(e.socketDest, msgC, errC, "network-error"), "re-connecting after network-error")

		case eventsubCloseCodeInvalidReconnect:
			e.logger.Warn("websocket reported invalid reconnect url")

		case websocket.CloseNormalClosure:
			// We don't take action here as a graceful close should
			// be initiated by us after establishing a new conn
			e.logger.Debug("websocket was closed normally")
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
		return socketInitialTimeout, errors.Wrap(err, "unmarshalling welcome message")
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
		if err := e.subscribe(); err != nil {
			return socketInitialTimeout, errors.Wrap(err, "subscribing to topics")
		}
	}

	e.logger.WithField("id", e.socketID).Debug("websocket connected successfully")

	// Configure proper keepalive
	return time.Duration(float64(payload.Session.KeepaliveTimeoutSeconds)*socketTimeoutGraceMultiplier) * time.Second, nil
}

func (e *EventSubSocketClient) subscribe() error {
	for _, st := range e.subscriptionTypes {
		if _, err := e.twitch.createEventSubSubscriptionWebsocket(context.Background(), eventSubSubscription{
			Type:      st.Event,
			Version:   st.Version,
			Condition: st.Condition,
			Transport: eventSubTransport{
				Method:    "websocket",
				SessionID: e.socketID,
			},
		}); err != nil {
			return errors.Wrapf(err, "subscribing to %s/%s", st.Event, st.Version)
		}

		e.logger.WithField("topic", strings.Join([]string{st.Event, st.Version}, "/")).Debug("subscribed to topic")
	}

	return nil
}

func (e eventSubSocketMessage) Unmarshal(dest any) error {
	return errors.Wrap(json.Unmarshal(e.Payload, dest), "unmarshalling payload")
}

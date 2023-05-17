package twitch

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	eventsubLiveSocketDest = "wss://eventsub.wss.twitch.tv/ws"
	socketInitialTimeout   = 10 * time.Second
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
			Condition map[string]string `json:"condition"`
			Transport struct {
				Method    string `json:"method"`
				SessionID string `json:"session_id"`
			} `json:"transport"`
			CreatedAt time.Time `json:"created_at"`
		} `json:"subscription"`
	}

	eventSubSocketPayloadSession struct {
		Session struct {
			ID                      string        `json:"id"`
			Status                  string        `json:"status"`
			ConnectedAt             time.Time     `json:"connected_at"`
			KeepaliveTimeoutSeconds time.Duration `json:"keepalive_timeout_seconds"`
			ReconnectURL            *string       `json:"reconnect_url"`
		} `json:"session"`
	}
)

func NewEventSubSocketClient(opts ...EventSubSocketClientOpt) (*EventSubSocketClient, error) {
	c := &EventSubSocketClient{}

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

func (e *EventSubSocketClient) Run() error {
	var (
		errC             = make(chan error, 10)
		keepaliveTimeout = socketInitialTimeout
		msgC             = make(chan eventSubSocketMessage, 1)
		socketTimeout    = time.NewTimer(keepaliveTimeout)
	)

	if err := e.connect(e.socketDest, msgC, errC); err != nil {
		return errors.Wrap(err, "establishing initial connection")
	}

	for {
		select {
		case err := <-errC:
			// Something went wrong
			var closeErr *websocket.CloseError
			if errors.As(err, &closeErr) {
				switch closeErr.Code {
				case eventsubCloseCodeInternalServerError:
					e.logger.Warn("websocket reported internal server error")
					if err = e.connect(e.socketDest, msgC, errC); err != nil {
						return errors.Wrap(err, "re-connecting after internal-server-error")
					}
					continue

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
					if err = e.connect(e.socketDest, msgC, errC); err != nil {
						return errors.Wrap(err, "re-connecting after network-timeout")
					}
					continue

				case eventsubCloseCodeNetworkError:
					e.logger.Warn("websocket reported network error")
					if err = e.connect(e.socketDest, msgC, errC); err != nil {
						return errors.Wrap(err, "re-connecting after network-error")
					}
					continue

				case eventsubCloseCodeInvalidReconnect:
					e.logger.Warn("websocket reported invalid reconnect url")

				case websocket.CloseNormalClosure:
					// We don't take action here as a graceful close should
					// be initiated by us after establishing a new conn
					e.logger.Debug("websocket was closed normally")
					continue

				default:
					// Some non-twitch close code we did not expect
					e.logger.WithError(closeErr).Error("websocket reported unexpected error code")
				}
			}

			return err

		case <-socketTimeout.C:
			// No message received, deeming connection dead
			socketTimeout.Reset(socketInitialTimeout)
			if err := e.connect(e.socketDest, msgC, errC); err != nil {
				errC <- errors.Wrap(err, "re-connecting after timeout")
				continue
			}

		case msg := <-msgC:
			// The keepalive timer is reset with each notification or
			// keepalive message.
			socketTimeout.Reset(keepaliveTimeout)

			switch msg.Metadata.MessageType {
			case eventsubSocketMessageTypeKeepalive:
				// Handle only for debug, timer reset is done above
				e.logger.Debug("keepalive received")

			case eventsubSocketMessageTypeNotification:
				// We got mail! Yay!
				e.logger.Warnf("Received message: %s", msg.Payload)

			case eventsubSocketMessageTypeReconnect:
				// Twitch politely asked us to reconnect
				var payload eventSubSocketPayloadSession
				if err := msg.Unmarshal(&payload); err != nil {
					errC <- errors.Wrap(err, "unmarshalling reconnect message")
					continue
				}
				if payload.Session.ReconnectURL == nil {
					errC <- errors.New("reconnect message did not contain reconnect_url")
					continue
				}
				if err := e.connect(*payload.Session.ReconnectURL, msgC, errC); err != nil {
					errC <- errors.Wrap(err, "re-connecting after reconnect message")
					continue
				}

			case eventsubSocketMessageTypeWelcome:
				var payload eventSubSocketPayloadSession
				if err := msg.Unmarshal(&payload); err != nil {
					errC <- errors.Wrap(err, "unmarshalling welcome message")
					continue
				}

				// Configure proper keepalive
				keepaliveTimeout = payload.Session.KeepaliveTimeoutSeconds * time.Second

				// Close old connection if present
				if e.conn != nil {
					// We had an existing connection, close it
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
					if err := e.subscribe(); err != nil {
						errC <- errors.Wrap(err, "subscribing to topics")
					}
				}

				e.socketID = payload.Session.ID
				e.logger.WithField("id", e.socketID).Debug("websocket connected successfully")

			default:
				e.logger.WithField("type", msg.Metadata.MessageType).Error("unknown message type received")
			}
		}
	}
}

func (e *EventSubSocketClient) connect(url string, msgC chan eventSubSocketMessage, errC chan error) error {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
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

func (e *EventSubSocketClient) subscribe() error {
	for _, st := range e.subscriptionTypes {
		if _, err := e.twitch.createEventSubSubscription(context.Background(), eventSubSubscription{
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
	}

	return nil
}

func (e eventSubSocketMessage) Unmarshal(dest any) error {
	return errors.Wrap(json.Unmarshal(e.Payload, dest), "unmarshalling payload")
}

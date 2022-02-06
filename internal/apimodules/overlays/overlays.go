package overlays

import (
	"embed"
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/Luzifer/twitch-bot/plugins"
)

const (
	authTimeout     = 10 * time.Second
	bufferSizeByte  = 1024
	socketKeepAlive = 5 * time.Second

	moduleUUID = "f9ca2b3a-baf6-45ea-a347-c626168665e8"

	msgTypeRequestAuth   = "_auth"
	msgTypeRequestReplay = "_replay"
)

type (
	storage struct {
		ChannelEvents map[string][]socketMessage `json:"channel_events"`

		lock sync.RWMutex
	}

	socketMessage struct {
		IsLive bool                     `json:"is_live"`
		Time   time.Time                `json:"time"`
		Type   string                   `json:"type"`
		Fields *plugins.FieldCollection `json:"fields"`
	}
)

var (
	//go:embed default/**
	embeddedOverlays embed.FS

	fsStack httpFSStack

	ptrStringEmpty = func(v string) *string { return &v }("")

	store          plugins.StorageManager
	storeExemption = []string{
		"join", "part", // Those make no sense for replay
	}
	storedObject = newStorage()

	subscribers     = map[string]func(event string, eventData *plugins.FieldCollection){}
	subscribersLock sync.RWMutex

	upgrader = websocket.Upgrader{
		ReadBufferSize:  bufferSizeByte,
		WriteBufferSize: bufferSizeByte,
	}

	validateToken plugins.ValidateTokenFunc
)

func Register(args plugins.RegistrationArguments) error {
	store = args.GetStorageManager()
	validateToken = args.ValidateToken

	args.RegisterAPIRoute(plugins.HTTPRouteRegistrationArgs{
		Description:  "Websocket subscriber for bot events",
		HandlerFunc:  handleSocketSubscription,
		Method:       http.MethodGet,
		Module:       "overlays",
		Name:         "Websocket",
		Path:         "/events.sock",
		ResponseType: plugins.HTTPRouteResponseTypeMultiple,
	})

	args.RegisterAPIRoute(plugins.HTTPRouteRegistrationArgs{
		HandlerFunc:       handleServeOverlayAsset,
		IsPrefix:          true,
		Method:            http.MethodGet,
		Module:            "overlays",
		Path:              "/",
		ResponseType:      plugins.HTTPRouteResponseTypeMultiple,
		SkipDocumentation: true,
	})

	args.RegisterEventHandler(func(event string, eventData *plugins.FieldCollection) error {
		subscribersLock.RLock()
		defer subscribersLock.RUnlock()

		for _, fn := range subscribers {
			fn(event, eventData)
		}

		if str.StringInSlice(event, storeExemption) {
			return nil
		}

		storedObject.AddEvent(plugins.DeriveChannel(nil, eventData), socketMessage{
			IsLive: false,
			Time:   time.Now(),
			Type:   event,
			Fields: eventData,
		})

		return errors.Wrap(
			store.SetModuleStore(moduleUUID, storedObject),
			"storing events database",
		)
	})

	fsStack = httpFSStack{
		newPrefixedFS("default", http.FS(embeddedOverlays)),
	}

	overlaysDir := os.Getenv("OVERLAYS_DIR")
	if ds, err := os.Stat(overlaysDir); err != nil || overlaysDir == "" || !ds.IsDir() {
		log.WithField("dir", overlaysDir).Warn("Overlays dir not specified, no dir or non existent")
	} else {
		fsStack = append(httpFSStack{http.Dir(overlaysDir)}, fsStack...)
	}

	return errors.Wrap(
		store.GetModuleStore(moduleUUID, storedObject),
		"loading module storage",
	)
}

func handleServeOverlayAsset(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix("/overlays", http.FileServer(fsStack)).ServeHTTP(w, r)
}

//nolint:funlen,gocognit,gocyclo // Not split in order to keep the socket logic in one place
func handleSocketSubscription(w http.ResponseWriter, r *http.Request) {
	var (
		connID = uuid.Must(uuid.NewV4()).String()
		logger = log.WithField("conn_id", connID)
	)

	// Upgrade connection to socket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.WithError(err).Error("Unable to upgrade socket")
		return
	}
	defer conn.Close()

	var (
		authTimeout  = time.NewTimer(authTimeout)
		connLock     = new(sync.Mutex)
		errC         = make(chan error, 1)
		isAuthorized bool
		sendMsgC     = make(chan socketMessage, 1)
	)

	// Register listener
	unsub := subscribeSocket(func(event string, eventData *plugins.FieldCollection) {
		sendMsgC <- socketMessage{
			IsLive: true,
			Time:   time.Now(),
			Type:   event,
			Fields: eventData,
		}
	})
	defer unsub()

	keepAlive := time.NewTicker(socketKeepAlive)
	defer keepAlive.Stop()
	go func() {
		for range keepAlive.C {
			connLock.Lock()

			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.WithError(err).Error("Unable to send ping message")
				connLock.Unlock()
				conn.Close()
				return
			}

			connLock.Unlock()
		}
	}()

	go func() {
		// Handle socket
		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				errC <- errors.Wrap(err, "reading from socket")
				return
			}

			switch messageType {
			case websocket.TextMessage:
				// This is fine and expected

			case websocket.BinaryMessage:
				// Wat?
				errC <- errors.New("binary message received")
				return

			case websocket.CloseMessage:
				// They want to go? Fine, have it that way!
				errC <- nil
				return

			default:
				logger.Debugf("Got unhandled message from socket: %d", messageType)
				continue
			}

			var recvMsg socketMessage
			if err = json.Unmarshal(p, &recvMsg); err != nil {
				errC <- errors.Wrap(err, "decoding message")
				return
			}

			if !isAuthorized && recvMsg.Type != msgTypeRequestAuth {
				// Socket is requesting stuff but is not authorized, we don't want them to be here!
				errC <- nil
				return
			}

			switch recvMsg.Type {
			case msgTypeRequestAuth:
				if err := validateToken(recvMsg.Fields.MustString("token", ptrStringEmpty), "overlays"); err != nil {
					errC <- errors.Wrap(err, "validating auth token")
					return
				}

				authTimeout.Stop()
				isAuthorized = true
				sendMsgC <- socketMessage{
					IsLive: true,
					Time:   time.Now(),
					Type:   msgTypeRequestAuth,
				}

			case msgTypeRequestReplay:
				go func() {
					for _, msg := range storedObject.GetChannelEvents(recvMsg.Fields.MustString("channel", ptrStringEmpty)) {
						sendMsgC <- msg
					}
				}()

			default:
				logger.WithField("type", recvMsg.Type).Warn("Got unexpected message type from frontend")
			}
		}
	}()

	for {
		select {
		case <-authTimeout.C:
			// Timeout was not stopped, no auth was done
			logger.Warn("socket failed to auth")
			return

		case err := <-errC:
			if err != nil {
				logger.WithError(err).Error("Message processing caused error")
			}
			return // We use nil-error to close the connection

		case msg := <-sendMsgC:
			if !isAuthorized {
				// Not authorized, we're silently dropping messages
				continue
			}

			connLock.Lock()
			if err := conn.WriteJSON(msg); err != nil {
				logger.WithError(err).Error("Unable to send socket message")
				connLock.Unlock()
				conn.Close()
			}
			connLock.Unlock()
		}
	}
}

func subscribeSocket(fn func(event string, eventData *plugins.FieldCollection)) func() {
	id := uuid.Must(uuid.NewV4()).String()

	subscribersLock.Lock()
	defer subscribersLock.Unlock()

	subscribers[id] = fn

	return func() { unsubscribeSocket(id) }
}

func unsubscribeSocket(id string) {
	subscribersLock.Lock()
	defer subscribersLock.Unlock()

	delete(subscribers, id)
}

// Storage

func newStorage() *storage {
	return &storage{
		ChannelEvents: make(map[string][]socketMessage),
	}
}

func (s *storage) AddEvent(channel string, evt socketMessage) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.ChannelEvents[channel] = append(s.ChannelEvents[channel], evt)
}

func (s *storage) GetChannelEvents(channel string) []socketMessage {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.ChannelEvents[channel]
}

// Implement marshaller interfaces
func (s *storage) MarshalStoredObject() ([]byte, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return json.Marshal(s)
}

func (s *storage) UnmarshalStoredObject(data []byte) error {
	if data == nil {
		// No data set yet, don't try to unmarshal
		return nil
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	return json.Unmarshal(data, s)
}

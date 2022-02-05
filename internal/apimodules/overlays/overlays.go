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
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/plugins"
)

const (
	bufferSizeByte  = 1024
	socketKeepAlive = 5 * time.Second
)

type (
	socketMessage struct {
		Type   string                 `json:"type"`
		Fields map[string]interface{} `json:"fields"`
	}
)

var (
	//go:embed default/**
	embeddedOverlays embed.FS

	fsStack httpFSStack

	subscribers     = map[string]func(event string, eventData *plugins.FieldCollection){}
	subscribersLock sync.RWMutex

	upgrader = websocket.Upgrader{
		ReadBufferSize:  bufferSizeByte,
		WriteBufferSize: bufferSizeByte,
	}
)

func Register(args plugins.RegistrationArguments) error {
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

		return nil
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

	return nil
}

func handleServeOverlayAsset(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix("/overlays", http.FileServer(fsStack)).ServeHTTP(w, r)
}

func handleSocketSubscription(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection to socket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithError(err).Error("Unable to upgrade socket")
		return
	}
	defer conn.Close()

	// Register listener
	connLock := new(sync.Mutex)

	unsub := subscribeSocket(func(event string, eventData *plugins.FieldCollection) {
		connLock.Lock()
		defer connLock.Unlock()

		if err := conn.WriteJSON(socketMessage{
			Type:   event,
			Fields: eventData.Data(),
		}); err != nil {
			log.WithError(err).Error("Unable to send socket message")
			connLock.Unlock()
			conn.Close()
		}
	})
	defer unsub()

	keepAlive := time.NewTicker(socketKeepAlive)
	defer keepAlive.Stop()
	go func() {
		for range keepAlive.C {
			connLock.Lock()

			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.WithError(err).Error("Unable to send ping message")
				connLock.Unlock()
				conn.Close()
				return
			}

			connLock.Unlock()
		}
	}()

	// Handle socket
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.WithError(err).Error("Unable to read from socket")
			return
		}

		switch messageType {
		case websocket.TextMessage:
			// This is fine and expected

		case websocket.BinaryMessage:
			// Wat?
			log.Warn("Got binary message from socket, disconnecting...")
			return

		case websocket.CloseMessage:
			// They want to go? Fine, have it that way!
			return

		default:
			log.Debugf("Got unhandled message from socket: %d", messageType)
			continue
		}

		var recvMsg socketMessage
		if err = json.Unmarshal(p, &recvMsg); err != nil {
			log.Warn("Got unreadable message from socket, disconnecting...")
			return
		}

		switch recvMsg.Type {
		default:
			log.WithField("type", recvMsg.Type).Warn("Got unexpected message type from frontend")
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

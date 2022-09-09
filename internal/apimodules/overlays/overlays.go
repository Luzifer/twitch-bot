package overlays

import (
	"embed"
	"encoding/json"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/Luzifer/twitch-bot/internal/database"
	"github.com/Luzifer/twitch-bot/plugins"
)

const (
	authTimeout     = 10 * time.Second
	bufferSizeByte  = 1024
	socketKeepAlive = 5 * time.Second

	moduleUUID = "f9ca2b3a-baf6-45ea-a347-c626168665e8"

	msgTypeRequestAuth = "_auth"
)

type (
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

	db database.Connector

	fsStack httpFSStack

	ptrStringEmpty = func(v string) *string { return &v }("")

	storeExemption = []string{
		"join", "part", // Those make no sense for replay
	}

	subscribers     = map[string]func(event string, eventData *plugins.FieldCollection){}
	subscribersLock sync.RWMutex

	upgrader = websocket.Upgrader{
		ReadBufferSize:  bufferSizeByte,
		WriteBufferSize: bufferSizeByte,
	}

	validateToken plugins.ValidateTokenFunc
)

func Register(args plugins.RegistrationArguments) error {
	db = args.GetDatabaseConnector()
	if err := db.Migrate("overlays", database.NewEmbedFSMigrator(schema, "schema")); err != nil {
		return errors.Wrap(err, "applying schema migration")
	}

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
		Description: "Fetch past events for the given channel",
		HandlerFunc: handleEventsReplay,
		Method:      http.MethodGet,
		Module:      "overlays",
		Name:        "Replay",
		Path:        "/events/{channel}",
		QueryParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "ISO / RFC3339 timestamp to fetch the events after",
				Name:        "since",
				Required:    false,
				Type:        "string",
			},
		},
		RequiresWriteAuth: true,
		ResponseType:      plugins.HTTPRouteResponseTypeJSON,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Channel to fetch the events from",
				Name:        "channel",
			},
		},
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

		return errors.Wrap(
			addEvent(plugins.DeriveChannel(nil, eventData), socketMessage{
				IsLive: false,
				Time:   time.Now(),
				Type:   event,
				Fields: eventData,
			}),
			"storing event",
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

	return nil
}

func handleEventsReplay(w http.ResponseWriter, r *http.Request) {
	var (
		channel = mux.Vars(r)["channel"]
		msgs    []socketMessage
		since   = time.Time{}
	)

	if s, err := time.Parse(time.RFC3339, r.URL.Query().Get("since")); err == nil {
		since = s
	}

	events, err := getChannelEvents("#" + strings.TrimLeft(channel, "#"))
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting channel events").Error(), http.StatusInternalServerError)
		return
	}

	for _, msg := range events {
		if msg.Time.Before(since) {
			continue
		}

		msgs = append(msgs, msg)
	}

	sort.Slice(msgs, func(i, j int) bool { return msgs[i].Time.Before(msgs[j].Time) })

	if err := json.NewEncoder(w).Encode(msgs); err != nil {
		http.Error(w, errors.Wrap(err, "encoding response").Error(), http.StatusInternalServerError)
		return
	}
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

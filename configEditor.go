package main

import (
	"embed"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/Luzifer/twitch-bot/twitch"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

const websocketPingInterval = 30 * time.Second

var (
	availableActorDocs     = []plugins.ActionDocumentation{}
	availableActorDocsLock sync.RWMutex

	//go:embed editor/*
	configEditorFrontend embed.FS

	upgrader = websocket.Upgrader{}
)

func registerActorDocumentation(doc plugins.ActionDocumentation) {
	availableActorDocsLock.Lock()
	defer availableActorDocsLock.Unlock()

	availableActorDocs = append(availableActorDocs, doc)
	sort.Slice(availableActorDocs, func(i, j int) bool {
		return availableActorDocs[i].Name < availableActorDocs[j].Name
	})
}

func init() {
	registerEditorAutoMessageRoutes()
	registerEditorFrontend()
	registerEditorGeneralConfigRoutes()
	registerEditorGlobalMethods()
	registerEditorRulesRoutes()
}

func registerEditorFrontend() {
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		f, err := configEditorFrontend.Open("editor/index.html")
		if err != nil {
			http.Error(w, errors.Wrap(err, "opening index.html").Error(), http.StatusNotFound)
			return
		}

		io.Copy(w, f)
	})

	router.HandleFunc("/editor/vars.json", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(struct {
			IRCBadges      []string
			KnownEvents    []*string
			TwitchClientID string
		}{
			IRCBadges:      twitch.KnownBadges,
			KnownEvents:    knownEvents,
			TwitchClientID: cfg.TwitchClient,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	router.PathPrefix("/editor").Handler(http.FileServer(http.FS(configEditorFrontend)))
}

package main

import (
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v2/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v2/plugins"
)

const websocketPingInterval = 30 * time.Second

var (
	availableActorDocs     = []plugins.ActionDocumentation{}
	availableActorDocsLock sync.RWMutex

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
			IRCBadges         []string
			KnownEvents       []*string
			TemplateFunctions []string
			TwitchClientID    string
			Version           string
		}{
			IRCBadges:         twitch.KnownBadges,
			KnownEvents:       knownEvents,
			TemplateFunctions: tplFuncs.GetFuncNames(),
			TwitchClientID:    cfg.TwitchClient,
			Version:           version,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	router.PathPrefix("/editor").Handler(http.FileServer(configEditorFrontend))
}

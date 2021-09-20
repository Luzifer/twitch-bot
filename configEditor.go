package main

import (
	"embed"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/gofrs/uuid/v3"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	availableActorDocs     = []plugins.ActionDocumentation{}
	availableActorDocsLock sync.RWMutex

	//go:embed editor/*
	configEditorFrontend embed.FS
)

func registerActorDocumentation(doc plugins.ActionDocumentation) {
	availableActorDocsLock.Lock()
	defer availableActorDocsLock.Unlock()

	availableActorDocs = append(availableActorDocs, doc)
}

type (
	configEditorGeneralConfig struct {
		BotEditors []string `json:"bot_editors"`
		Channels   []string `json:"channels"`
	}
)

func init() {
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
			TwitchClientID string
		}{
			TwitchClientID: cfg.TwitchClient,
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	router.PathPrefix("/editor").Handler(http.FileServer(http.FS(configEditorFrontend)))

	for _, rd := range []plugins.HTTPRouteRegistrationArgs{
		{
			Description: "Returns the documentation for available actions",
			HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
				availableActorDocsLock.Lock()
				defer availableActorDocsLock.Unlock()

				if err := json.NewEncoder(w).Encode(availableActorDocs); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			},
			Method:       http.MethodGet,
			Module:       "config-editor",
			Name:         "Get available actions",
			Path:         "/actions",
			ResponseType: plugins.HTTPRouteResponseTypeJSON,
		},
		{
			Description: "Returns the current set of configured auto-messages in JSON format",
			HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewEncoder(w).Encode(config.AutoMessages); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			},
			Method:              http.MethodGet,
			Module:              "config-editor",
			Name:                "Get current auto-messages",
			Path:                "/auto-messages",
			RequiresEditorsAuth: true,
			ResponseType:        plugins.HTTPRouteResponseTypeJSON,
		},
		{
			Description: "Adds a new Auto-Message",
			HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
				msg := &autoMessage{}
				if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				msg.UUID = uuid.Must(uuid.NewV4()).String()

				if err := patchConfig(cfg.Config, func(c *configFile) error {
					c.AutoMessages = append(c.AutoMessages, msg)
					return nil
				}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusCreated)
			},
			Method:              http.MethodPost,
			Module:              "config-editor",
			Name:                "Add Auto-Message",
			Path:                "/auto-messages",
			RequiresEditorsAuth: true,
			ResponseType:        plugins.HTTPRouteResponseTypeTextPlain,
		},
		{
			Description: "Deletes the given Auto-Message",
			HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if err := patchConfig(cfg.Config, func(c *configFile) error {
					var (
						id  = mux.Vars(r)["uuid"]
						tmp []*autoMessage
					)

					for i := range c.AutoMessages {
						if c.AutoMessages[i].ID() == id {
							continue
						}
						tmp = append(tmp, c.AutoMessages[i])
					}

					c.AutoMessages = tmp

					return nil
				}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusNoContent)
			},
			Method:              http.MethodDelete,
			Module:              "config-editor",
			Name:                "Delete Auto-Message",
			Path:                "/auto-messages/{uuid}",
			RequiresEditorsAuth: true,
			ResponseType:        plugins.HTTPRouteResponseTypeTextPlain,
			RouteParams: []plugins.HTTPRouteParamDocumentation{
				{
					Description: "UUID of the auto-message to delete",
					Name:        "uuid",
					Required:    false,
					Type:        "string",
				},
			},
		},
		{
			Description: "Updates the given Auto-Message",
			HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
				msg := &autoMessage{}
				if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				if err := patchConfig(cfg.Config, func(c *configFile) error {
					id := mux.Vars(r)["uuid"]

					for i := range c.AutoMessages {
						if c.AutoMessages[i].ID() == id {
							c.AutoMessages[i] = msg
						}
					}

					return nil
				}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusNoContent)
			},
			Method:              http.MethodPut,
			Module:              "config-editor",
			Name:                "Update Auto-Message",
			Path:                "/auto-messages/{uuid}",
			RequiresEditorsAuth: true,
			ResponseType:        plugins.HTTPRouteResponseTypeTextPlain,
			RouteParams: []plugins.HTTPRouteParamDocumentation{
				{
					Description: "UUID of the auto-message to delete",
					Name:        "uuid",
					Required:    false,
					Type:        "string",
				},
			},
		},
		{
			Description: "Returns the current general config",
			HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewEncoder(w).Encode(configEditorGeneralConfig{
					BotEditors: config.BotEditors,
					Channels:   config.Channels,
				}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			},
			Method:              http.MethodGet,
			Module:              "config-editor",
			Name:                "Get general config",
			Path:                "/general",
			RequiresEditorsAuth: true,
			ResponseType:        plugins.HTTPRouteResponseTypeJSON,
		},
		{
			Description: "Updates the general config",
			HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
				var payload configEditorGeneralConfig

				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				for i := range payload.BotEditors {
					usr, err := twitchClient.GetUserInformation(payload.BotEditors[i])
					if err != nil {
						http.Error(w, errors.Wrap(err, "getting bot editor profile").Error(), http.StatusInternalServerError)
						return
					}

					payload.BotEditors[i] = usr.ID
				}

				if err := patchConfig(cfg.Config, func(cfg *configFile) error {
					cfg.Channels = payload.Channels
					cfg.BotEditors = payload.BotEditors

					return nil
				}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusNoContent)
			},
			Method:              http.MethodPut,
			Module:              "config-editor",
			Name:                "Update general config",
			Path:                "/general",
			RequiresEditorsAuth: true,
			ResponseType:        plugins.HTTPRouteResponseTypeTextPlain,
		},
		{
			Description: "Returns the current set of configured rules in JSON format",
			HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewEncoder(w).Encode(config.Rules); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			},
			Method:              http.MethodGet,
			Module:              "config-editor",
			Name:                "Get current rules",
			Path:                "/rules",
			RequiresEditorsAuth: true,
			ResponseType:        plugins.HTTPRouteResponseTypeJSON,
		},
		{
			Description: "Returns information about a Twitch user to properly display bot editors",
			HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
				usr, err := twitchClient.GetUserInformation(r.FormValue("user"))
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if err := json.NewEncoder(w).Encode(usr); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			},
			Method: http.MethodGet,
			Module: "config-editor",
			Name:   "Get user information",
			Path:   "/user",
			QueryParams: []plugins.HTTPRouteParamDocumentation{
				{
					Description: "The user to query the information for",
					Name:        "user",
					Required:    true,
					Type:        "string",
				},
			},
			RequiresEditorsAuth: true,
			ResponseType:        plugins.HTTPRouteResponseTypeJSON,
		},
		{
			Description: "Validate a cron expression and return the next executions",
			HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
				sched, err := cronParser.Parse(r.FormValue("cron"))
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				var (
					lt  = time.Now()
					out []time.Time
				)

				if id := r.FormValue("uuid"); id != "" {
					for _, a := range config.AutoMessages {
						if a.ID() != id {
							continue
						}
						lt = a.lastMessageSent
						break
					}
				}

				for i := 0; i < 3; i++ {
					lt = sched.Next(lt)
					out = append(out, lt)
				}

				if err := json.NewEncoder(w).Encode(out); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			},
			Method: http.MethodPut,
			Module: "config-editor",
			Name:   "Validate cron expression",
			Path:   "/validate-cron",
			QueryParams: []plugins.HTTPRouteParamDocumentation{
				{
					Description: "The cron expression to test",
					Name:        "cron",
					Required:    true,
					Type:        "string",
				},
				{
					Description: "Check cron with last execution of auto-message",
					Name:        "uuid",
					Required:    false,
					Type:        "string",
				},
			},
			ResponseType: plugins.HTTPRouteResponseTypeJSON,
		},
		{
			Description: "Validate a regular expression against the RE2 regex parser",
			HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if _, err := regexp.Compile(r.FormValue("regexp")); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				w.WriteHeader(http.StatusNoContent)
			},
			Method: http.MethodPut,
			Module: "config-editor",
			Name:   "Validate regular expression",
			Path:   "/validate-regex",
			QueryParams: []plugins.HTTPRouteParamDocumentation{
				{
					Description: "The regular expression to test",
					Name:        "regexp",
					Required:    true,
					Type:        "string",
				},
			},
			ResponseType: plugins.HTTPRouteResponseTypeTextPlain,
		},
	} {
		if err := registerRoute(rd); err != nil {
			log.WithError(err).Fatal("Unable to register config editor route")
		}
	}
}

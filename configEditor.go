package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/Luzifer/twitch-bot/plugins"
	log "github.com/sirupsen/logrus"
)

var (
	availableActorDocs     = []plugins.ActionDocumentation{}
	availableActorDocsLock sync.RWMutex
)

func registerActorDocumentation(doc plugins.ActionDocumentation) {
	availableActorDocsLock.Lock()
	defer availableActorDocsLock.Unlock()

	availableActorDocs = append(availableActorDocs, doc)
}

func init() {
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
			Description: "Returns the current set of configured rules in JSON format",
			HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewEncoder(w).Encode(config.Rules); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			},
			Method:       http.MethodGet,
			Module:       "config-editor",
			Name:         "Get current configuration",
			Path:         "/rules",
			ResponseType: plugins.HTTPRouteResponseTypeJSON,
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

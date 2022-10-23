package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v2/plugins"
)

func registerEditorGlobalMethods() {
	for _, rd := range []plugins.HTTPRouteRegistrationArgs{
		{
			Description:  "Returns the documentation for available actions",
			HandlerFunc:  configEditorGlobalGetActions,
			Method:       http.MethodGet,
			Module:       "config-editor",
			Name:         "Get available actions",
			Path:         "/actions",
			ResponseType: plugins.HTTPRouteResponseTypeJSON,
		},
		{
			Description:  "Returns all available modules for auth",
			HandlerFunc:  configEditorGlobalGetModules,
			Method:       http.MethodGet,
			Module:       "config-editor",
			Name:         "Get available modules",
			Path:         "/modules",
			ResponseType: plugins.HTTPRouteResponseTypeJSON,
		},
		{
			Description: "Returns information about a Twitch user to properly display bot editors",
			HandlerFunc: configEditorGlobalGetUser,
			Method:      http.MethodGet,
			Module:      "config-editor",
			Name:        "Get user information",
			Path:        "/user",
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
			Description:  "Subscribe for configuration changes",
			HandlerFunc:  configEditorGlobalSubscribe,
			Method:       http.MethodGet,
			Module:       "config-editor",
			Name:         "Websocket: Subscribe config changes",
			Path:         "/notify-config",
			ResponseType: plugins.HTTPRouteResponseTypeTextPlain,
		},
		{
			Description: "Validate a cron expression and return the next executions",
			HandlerFunc: configEditorGlobalValidateCron,
			Method:      http.MethodPut,
			Module:      "config-editor",
			Name:        "Validate cron expression",
			Path:        "/validate-cron",
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
			HandlerFunc: configEditorGlobalValidateRegex,
			Method:      http.MethodPut,
			Module:      "config-editor",
			Name:        "Validate regular expression",
			Path:        "/validate-regex",
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

func configEditorGlobalGetActions(w http.ResponseWriter, r *http.Request) {
	availableActorDocsLock.Lock()
	defer availableActorDocsLock.Unlock()

	if err := json.NewEncoder(w).Encode(availableActorDocs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func configEditorGlobalGetModules(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(knownModules); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func configEditorGlobalGetUser(w http.ResponseWriter, r *http.Request) {
	usr, err := twitchClient.GetUserInformation(r.FormValue("user"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(usr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func configEditorGlobalSubscribe(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithError(err).Error("Unable to initialize websocket")
		return
	}
	defer conn.Close()

	var (
		configReloadNotify = make(chan struct{}, 1)
		pingTimer          = time.NewTicker(websocketPingInterval)
		unsubscribe        = registerConfigReloadHook(func() { configReloadNotify <- struct{}{} })
	)
	defer unsubscribe()

	type socketMsg struct {
		MsgType string `json:"msg_type"`
	}

	for {
		select {
		case <-configReloadNotify:
			if err := conn.WriteJSON(socketMsg{
				MsgType: "configReload",
			}); err != nil {
				log.WithError(err).Debug("Unable to send websocket notification")
				return
			}

		case <-pingTimer.C:
			if err := conn.WriteJSON(socketMsg{
				MsgType: "ping",
			}); err != nil {
				log.WithError(err).Debug("Unable to send websocket ping")
				return
			}

		}
	}
}

func configEditorGlobalValidateCron(w http.ResponseWriter, r *http.Request) {
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
}

func configEditorGlobalValidateRegex(w http.ResponseWriter, r *http.Request) {
	if _, err := regexp.Compile(r.FormValue("regexp")); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

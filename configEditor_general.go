package main

import (
	"encoding/json"
	"net/http"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type (
	configEditorGeneralConfig struct {
		BotEditors []string `json:"bot_editors"`
		Channels   []string `json:"channels"`
	}
)

func registerEditorGeneralConfigRoutes() {
	for _, rd := range []plugins.HTTPRouteRegistrationArgs{
		{
			Description:         "Returns the current general config",
			HandlerFunc:         configEditorHandleGeneralGet,
			Method:              http.MethodGet,
			Module:              "config-editor",
			Name:                "Get general config",
			Path:                "/general",
			RequiresEditorsAuth: true,
			ResponseType:        plugins.HTTPRouteResponseTypeJSON,
		},
		{
			Description:         "Updates the general config",
			HandlerFunc:         configEditorHandleGeneralUpdate,
			Method:              http.MethodPut,
			Module:              "config-editor",
			Name:                "Update general config",
			Path:                "/general",
			RequiresEditorsAuth: true,
			ResponseType:        plugins.HTTPRouteResponseTypeTextPlain,
		},
	} {
		if err := registerRoute(rd); err != nil {
			log.WithError(err).Fatal("Unable to register config editor route")
		}
	}
}

func configEditorHandleGeneralGet(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(configEditorGeneralConfig{
		BotEditors: config.BotEditors,
		Channels:   config.Channels,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func configEditorHandleGeneralUpdate(w http.ResponseWriter, r *http.Request) {
	user, _, err := getAuthorizationFromRequest(r)
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting authorized user").Error(), http.StatusInternalServerError)
	}

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

	if err := patchConfig(cfg.Config, user, "", "Update general config", func(cfg *configFile) error {
		cfg.Channels = payload.Channels
		cfg.BotEditors = payload.BotEditors

		return nil
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofrs/uuid/v3"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v2/plugins"
)

type (
	configEditorGeneralConfig struct {
		BotEditors       []string        `json:"bot_editors"`
		BotName          *string         `json:"bot_name,omitempty"`
		Channels         []string        `json:"channels"`
		ChannelHasScopes map[string]bool `json:"channel_has_scopes"`
	}
)

func registerEditorGeneralConfigRoutes() {
	for _, rd := range []plugins.HTTPRouteRegistrationArgs{
		{
			Description:         "Add new authorization token",
			HandlerFunc:         configEditorHandleGeneralAddAuthToken,
			Method:              http.MethodPost,
			Module:              "config-editor",
			Name:                "Add authorization token",
			Path:                "/auth-tokens",
			RequiresEditorsAuth: true,
			ResponseType:        plugins.HTTPRouteResponseTypeJSON,
		},
		{
			Description:         "Delete authorization token",
			HandlerFunc:         configEditorHandleGeneralDeleteAuthToken,
			Method:              http.MethodDelete,
			Module:              "config-editor",
			Name:                "Delete authorization token",
			Path:                "/auth-tokens/{handle}",
			RequiresEditorsAuth: true,
			ResponseType:        plugins.HTTPRouteResponseTypeTextPlain,
			RouteParams: []plugins.HTTPRouteParamDocumentation{
				{
					Description: "UUID of the auth-token to delete",
					Name:        "handle",
					Required:    true,
					Type:        "string",
				},
			},
		},
		{
			Description:         "List authorization tokens",
			HandlerFunc:         configEditorHandleGeneralListAuthTokens,
			Method:              http.MethodGet,
			Module:              "config-editor",
			Name:                "List authorization tokens",
			Path:                "/auth-tokens",
			RequiresEditorsAuth: true,
			ResponseType:        plugins.HTTPRouteResponseTypeJSON,
		},
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
		{
			Description:         "Get Bot-Auth URLs for updating bot token and channel scopes",
			HandlerFunc:         configEditorHandleGeneralAuthURLs,
			Method:              http.MethodGet,
			Module:              "config-editor",
			Name:                "Get Bot-Auth-URLs",
			Path:                "/auth-urls",
			RequiresEditorsAuth: true,
			ResponseType:        plugins.HTTPRouteResponseTypeJSON,
		},
	} {
		if err := registerRoute(rd); err != nil {
			log.WithError(err).Fatal("Unable to register config editor route")
		}
	}
}

func configEditorHandleGeneralAddAuthToken(w http.ResponseWriter, r *http.Request) {
	user, _, err := getAuthorizationFromRequest(r)
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting authorized user").Error(), http.StatusInternalServerError)
		return
	}

	var payload configAuthToken
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, errors.Wrap(err, "reading payload").Error(), http.StatusBadRequest)
		return
	}

	if err = fillAuthToken(&payload); err != nil {
		http.Error(w, errors.Wrap(err, "hashing token").Error(), http.StatusInternalServerError)
		return
	}

	if err := patchConfig(cfg.Config, user, "", "Add auth-token", func(cfg *configFile) error {
		cfg.AuthTokens[uuid.Must(uuid.NewV4()).String()] = payload
		return nil
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func configEditorHandleGeneralAuthURLs(w http.ResponseWriter, r *http.Request) {
	var out struct {
		UpdateBotToken      string `json:"update_bot_token"`
		UpdateChannelScopes string `json:"update_channel_scopes"`
	}

	params := make(url.Values)
	params.Set("client_id", cfg.TwitchClient)
	params.Set("redirect_uri", strings.Join([]string{
		strings.TrimRight(cfg.BaseURL, "/"),
		"auth", "update-bot-token",
	}, "/"))
	params.Set("response_type", "code")
	params.Set("scope", strings.Join(botDefaultScopes, " "))
	params.Set("state", instanceState)

	out.UpdateBotToken = fmt.Sprintf("https://id.twitch.tv/oauth2/authorize?%s", params.Encode())

	params.Set("redirect_uri", strings.Join([]string{
		strings.TrimRight(cfg.BaseURL, "/"),
		"auth", "update-channel-scopes",
	}, "/"))
	params.Set("scope", strings.Join(channelDefaultScopes, " "))

	out.UpdateChannelScopes = fmt.Sprintf("https://id.twitch.tv/oauth2/authorize?%s", params.Encode())

	if err := json.NewEncoder(w).Encode(out); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func configEditorHandleGeneralDeleteAuthToken(w http.ResponseWriter, r *http.Request) {
	user, _, err := getAuthorizationFromRequest(r)
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting authorized user").Error(), http.StatusInternalServerError)
	}

	if err := patchConfig(cfg.Config, user, "", "Delete auth-token", func(cfg *configFile) error {
		delete(cfg.AuthTokens, mux.Vars(r)["handle"])

		return nil
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func configEditorHandleGeneralGet(w http.ResponseWriter, r *http.Request) {
	var (
		elevated = make(map[string]bool)
		err      error
	)

	for _, ch := range config.Channels {
		if elevated[ch], err = accessService.HasPermissionsForChannel(ch, channelDefaultScopes...); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	var uName *string
	if n, err := twitchClient.GetAuthorizedUsername(); err == nil {
		uName = &n
	}

	if err := json.NewEncoder(w).Encode(configEditorGeneralConfig{
		BotEditors:       config.BotEditors,
		BotName:          uName,
		Channels:         config.Channels,
		ChannelHasScopes: elevated,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func configEditorHandleGeneralListAuthTokens(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(config.AuthTokens); err != nil {
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

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofrs/uuid/v3"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/Luzifer/twitch-bot/twitch"
)

var instanceState = uuid.Must(uuid.NewV4()).String()

func init() {
	for _, rd := range []plugins.HTTPRouteRegistrationArgs{
		{
			Description:  "Updates the bots token for connection to chat and API",
			HandlerFunc:  handleAuthUpdateBotToken,
			Method:       http.MethodGet,
			Module:       "auth",
			Name:         "Update bot token",
			Path:         "/update-bot-token",
			ResponseType: plugins.HTTPRouteResponseTypeTextPlain,
		},
		{
			Description:  "Updates scope configuration for EventSub subscription of a channel",
			HandlerFunc:  handleAuthUpdateChannelGrant,
			Method:       http.MethodGet,
			Module:       "auth",
			Name:         "Update channel scopes",
			Path:         "/update-channel-scopes",
			ResponseType: plugins.HTTPRouteResponseTypeTextPlain,
		},
	} {
		if err := registerRoute(rd); err != nil {
			log.WithError(err).Fatal("Unable to register auth routes")
		}
	}
}

func handleAuthUpdateBotToken(w http.ResponseWriter, r *http.Request) {
	var (
		code  = r.FormValue("code")
		state = r.FormValue("state")
	)

	if state != instanceState {
		http.Error(w, "invalid state, please start again", http.StatusBadRequest)
		return
	}

	params := make(url.Values)
	params.Set("client_id", cfg.TwitchClient)
	params.Set("client_secret", cfg.TwitchClientSecret)
	params.Set("code", code)
	params.Set("grant_type", "authorization_code")
	params.Set("redirect_uri", strings.Join([]string{
		strings.TrimRight(cfg.BaseURL, "/"),
		"auth", "update-bot-token",
	}, "/"))

	req, _ := http.NewRequestWithContext(r.Context(), http.MethodPost, fmt.Sprintf("https://id.twitch.tv/oauth2/token?%s", params.Encode()), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting access token").Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var rData twitch.OAuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&rData); err != nil {
		http.Error(w, errors.Wrap(err, "decoding access token").Error(), http.StatusInternalServerError)
		return
	}

	botUser, err := twitch.New(cfg.TwitchClient, cfg.TwitchClientSecret, rData.AccessToken, "").GetAuthorizedUsername()
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting authorized user").Error(), http.StatusInternalServerError)
		return
	}

	if err = store.UpdateBotToken(rData.AccessToken, rData.RefreshToken); err != nil {
		http.Error(w, errors.Wrap(err, "storing access token").Error(), http.StatusInternalServerError)
		return
	}

	twitchClient.UpdateToken(rData.AccessToken, rData.RefreshToken)

	if err = store.SetExtendedPermissions(botUser, storageExtendedPermission{
		AccessToken:  rData.AccessToken,
		RefreshToken: rData.RefreshToken,
		Scopes:       rData.Scope,
	}, true); err != nil {
		http.Error(w, errors.Wrap(err, "storing access scopes").Error(), http.StatusInternalServerError)
		return
	}

	http.Error(w, fmt.Sprintf("Authorization as %q complete, you can now close this window.", botUser), http.StatusOK)
}

func handleAuthUpdateChannelGrant(w http.ResponseWriter, r *http.Request) {
	var (
		code  = r.FormValue("code")
		state = r.FormValue("state")
	)

	if state != instanceState {
		http.Error(w, "invalid state, please start again", http.StatusBadRequest)
		return
	}

	params := make(url.Values)
	params.Set("client_id", cfg.TwitchClient)
	params.Set("client_secret", cfg.TwitchClientSecret)
	params.Set("code", code)
	params.Set("grant_type", "authorization_code")
	params.Set("redirect_uri", strings.Join([]string{
		strings.TrimRight(cfg.BaseURL, "/"),
		"auth", "update-channel-scopes",
	}, "/"))

	req, _ := http.NewRequestWithContext(r.Context(), http.MethodPost, fmt.Sprintf("https://id.twitch.tv/oauth2/token?%s", params.Encode()), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting access token").Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var rData twitch.OAuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&rData); err != nil {
		http.Error(w, errors.Wrap(err, "decoding access token").Error(), http.StatusInternalServerError)
		return
	}

	grantUser, err := twitch.New(cfg.TwitchClient, cfg.TwitchClientSecret, rData.AccessToken, "").GetAuthorizedUsername()
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting authorized user").Error(), http.StatusInternalServerError)
		return
	}

	if err = store.SetExtendedPermissions(grantUser, storageExtendedPermission{
		AccessToken:  rData.AccessToken,
		RefreshToken: rData.RefreshToken,
		Scopes:       rData.Scope,
	}, false); err != nil {
		http.Error(w, errors.Wrap(err, "storing access token").Error(), http.StatusInternalServerError)
		return
	}

	http.Error(w, fmt.Sprintf("Scopes for %q updated, you can now close this window.", grantUser), http.StatusOK)
}

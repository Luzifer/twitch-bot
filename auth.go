package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofrs/uuid/v3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
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
			logrus.WithError(err).Fatal("Unable to register auth routes")
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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Error("closing response body (leaked fd)")
		}
	}()

	var rData twitch.OAuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&rData); err != nil {
		http.Error(w, errors.Wrap(err, "decoding access token").Error(), http.StatusInternalServerError)
		return
	}

	_, botUser, err := twitch.New(cfg.TwitchClient, cfg.TwitchClientSecret, rData.AccessToken, "").GetAuthorizedUser(r.Context())
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting authorized user").Error(), http.StatusInternalServerError)
		return
	}

	if err = accessService.SetBotUsername(botUser); err != nil {
		http.Error(w, errors.Wrap(err, "storing bot username").Error(), http.StatusInternalServerError)
		return
	}

	twitchClient.UpdateToken(rData.AccessToken, rData.RefreshToken)

	if err = accessService.SetExtendedTwitchCredentials(botUser, rData.AccessToken, rData.RefreshToken, rData.Scope); err != nil {
		http.Error(w, errors.Wrap(err, "storing access scopes").Error(), http.StatusInternalServerError)
		return
	}

	http.Error(w, fmt.Sprintf("Authorization as %q complete, you can now close this window.", botUser), http.StatusOK)

	frontendNotifyHooks.Ping(frontendNotifyTypeReload) // Tell frontend to update its config
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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Error("closing response body (leaked fd)")
		}
	}()

	var rData twitch.OAuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&rData); err != nil {
		http.Error(w, errors.Wrap(err, "decoding access token").Error(), http.StatusInternalServerError)
		return
	}

	_, grantUser, err := twitch.New(cfg.TwitchClient, cfg.TwitchClientSecret, rData.AccessToken, "").GetAuthorizedUser(r.Context())
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting authorized user").Error(), http.StatusInternalServerError)
		return
	}

	if err = accessService.SetExtendedTwitchCredentials(grantUser, rData.AccessToken, rData.RefreshToken, rData.Scope); err != nil {
		http.Error(w, errors.Wrap(err, "storing access token").Error(), http.StatusInternalServerError)
		return
	}

	http.Error(w, fmt.Sprintf("Scopes for %q updated, you can now close this window.", grantUser), http.StatusOK)

	frontendNotifyHooks.Ping(frontendNotifyTypeReload) // Tell frontend to update its config
}

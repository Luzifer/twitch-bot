package main

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
)

func getAuthorizationFromRequest(r *http.Request) (string, *twitch.Client, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return "", nil, errors.New("no authorization provided")
	}

	tc := twitch.New(cfg.TwitchClient, cfg.TwitchClientSecret, token, "")

	_, user, err := tc.GetAuthorizedUser()
	return user, tc, errors.Wrap(err, "getting authorized user")
}

func botEditorAuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, tc, err := getAuthorizationFromRequest(r)
		if err != nil {
			http.Error(w, errors.Wrap(err, "getting authorized user").Error(), http.StatusForbidden)
			return
		}

		id, err := tc.GetIDForUsername(user)
		if err != nil {
			http.Error(w, errors.Wrap(err, "getting ID for authorized user").Error(), http.StatusForbidden)
			return
		}

		if !str.StringInSlice(user, config.BotEditors) && !str.StringInSlice(id, config.BotEditors) {
			http.Error(w, "user is not authorized", http.StatusForbidden)
			return
		}

		h.ServeHTTP(w, r)
	})
}

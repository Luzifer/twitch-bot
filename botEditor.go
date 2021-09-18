package main

import (
	"net/http"

	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/Luzifer/twitch-bot/twitch"
	"github.com/pkg/errors"
)

const twitchClientID = ""

func botEditorAuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "no auth-token provided", http.StatusForbidden)
			return
		}

		tc := twitch.New(twitchClientID, token)

		user, err := tc.GetAuthorizedUsername()
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

package main

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
)

func getAuthorizationFromRequest(r *http.Request) (string, *twitch.Client, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return "", nil, errors.New("no authorization provided")
	}

	tc := twitch.New(cfg.TwitchClient, cfg.TwitchClientSecret, token, "")

	_, user, err := tc.GetAuthorizedUser(r.Context())
	return user, tc, errors.Wrap(err, "getting authorized user")
}

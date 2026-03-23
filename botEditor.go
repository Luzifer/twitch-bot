package main

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
)

func getAuthorizedUserFromRequest(r *http.Request) (string, *twitch.Client, error) {
	tokenType, token, ok := getAuthorizationTokenFromRequest(r)
	if !ok {
		return "", nil, errors.New("no authorization provided")
	}

	if tokenType == authTokenTypeInternal {
		return "", nil, errors.New("internal tokens are not supported for Twitch user lookup")
	}

	tc := twitch.New(cfg.TwitchClient, cfg.TwitchClientSecret, token, "")

	_, user, err := tc.GetAuthorizedUser(r.Context())
	return user, tc, errors.Wrap(err, "getting authorized user")
}

package main

import (
	"errors"
	"fmt"
	"net/http"

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
	if err != nil {
		return user, tc, fmt.Errorf("getting authorized user: %w", err)
	}

	return user, tc, nil
}

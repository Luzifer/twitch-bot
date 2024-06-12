package main

import (
	"context"
	"net/http"
	"time"

	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/Luzifer/twitch-bot/v3/internal/service/authcache"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/pkg/errors"
)

const internalTokenAuthCacheExpiry = 5 * time.Minute

func authBackendInternalAppToken(token string) (modules []string, expiresAt time.Time, err error) {
	for _, auth := range config.AuthTokens {
		if auth.validate(token) != nil {
			continue
		}

		// We found a matching token
		return auth.Modules, time.Now().Add(internalTokenAuthCacheExpiry), nil
	}

	return nil, time.Time{}, authcache.ErrUnauthorized
}

func authBackendInternalEditorToken(token string) ([]string, time.Time, error) {
	id, user, expiresAt, err := editorTokenService.ValidateLoginToken(token)
	if err != nil {
		// None of our tokens: Nay.
		return nil, time.Time{}, authcache.ErrUnauthorized
	}

	if !str.StringInSlice(user, config.BotEditors) && !str.StringInSlice(id, config.BotEditors) {
		// That user is none of our editors: Deny access
		return nil, time.Time{}, authcache.ErrUnauthorized
	}

	// Editors have full access: Return module "*"
	return []string{"*"}, expiresAt, nil
}

func authBackendTwitchToken(token string) (modules []string, expiresAt time.Time, err error) {
	tc := twitch.New(cfg.TwitchClient, cfg.TwitchClientSecret, token, "")

	var httpError twitch.HTTPError

	id, user, err := tc.GetAuthorizedUser(context.Background())
	switch {
	case err == nil:
		// We got a valid user, continue check below
		if !str.StringInSlice(user, config.BotEditors) && !str.StringInSlice(id, config.BotEditors) {
			// That user is none of our editors: Deny access
			return nil, time.Time{}, authcache.ErrUnauthorized
		}

		_, _, expiresAt, err = tc.GetTokenInfo(context.Background())
		if err != nil {
			return nil, time.Time{}, errors.Wrap(err, "getting token expiry")
		}

		// Editors have full access: Return module "*"
		return []string{"*"}, expiresAt, nil

	case errors.As(err, &httpError):
		// We either got "forbidden" or we got another error
		if httpError.Code == http.StatusUnauthorized {
			// That token wasn't valid or not a Twitch token: Unauthorized
			return nil, time.Time{}, authcache.ErrUnauthorized
		}

		return nil, time.Time{}, errors.Wrap(err, "validating Twitch token")

	default:
		// Something else went wrong
		return nil, time.Time{}, errors.Wrap(err, "validating Twitch token")
	}
}

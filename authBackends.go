package main

import (
	"time"

	"github.com/Luzifer/twitch-bot/v3/internal/service/authcache"
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
	_, _, expiresAt, modules, err := editorTokenService.ValidateLoginToken(token)
	if err != nil {
		// None of our tokens: Nay.
		return nil, time.Time{}, authcache.ErrUnauthorized
	}

	return modules, expiresAt, nil
}

package main

import (
	"time"

	"github.com/Luzifer/go_helpers/v2/str"
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

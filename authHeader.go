package main

import (
	"net/http"
	"strings"

	"github.com/Luzifer/twitch-bot/v3/internal/service/authcache"
)

type authTokenType uint8

const (
	authTokenTypeUnknown authTokenType = iota
	authTokenTypeLegacy
	authTokenTypeInternal
	authTokenTypeTwitch
)

func getAuthorizationTokenFromRequest(r *http.Request) (tokenType authTokenType, token string, ok bool) {
	_, pass, hasBasicAuth := r.BasicAuth()
	switch {
	case hasBasicAuth && pass != "":
		return authTokenTypeLegacy, pass, true

	case r.Header.Get("Authorization") == "":
		return authTokenTypeUnknown, "", false
	}

	authHeader := r.Header.Get("Authorization")
	authType, token, hadPrefix := strings.Cut(authHeader, " ")
	switch {
	case !hadPrefix:
		// Legacy: Accept `Authorization: tokenhere`
		return authTokenTypeLegacy, authType, true

	case strings.EqualFold(authType, "token"):
		return authTokenTypeInternal, token, true

	case strings.EqualFold(authType, "twitch"):
		return authTokenTypeTwitch, token, true

	default:
		return authTokenTypeUnknown, "", false
	}
}

func (a authTokenType) Backend() string {
	switch a {
	case authTokenTypeInternal:
		return "internal"
	case authTokenTypeTwitch:
		return "twitch"

	default:
		return authcache.AuthBackendAny
	}
}

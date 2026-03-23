package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetAuthorizationTokenFromRequest(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		setupReq  func(*http.Request)
		tokenType authTokenType
		token     string
		ok        bool
	}{
		{
			name: "legacy raw header",
			setupReq: func(r *http.Request) {
				r.Header.Set("Authorization", "raw-token")
			},
			tokenType: authTokenTypeLegacy,
			token:     "raw-token",
			ok:        true,
		},
		{
			name: "internal prefixed header",
			setupReq: func(r *http.Request) {
				r.Header.Set("Authorization", "Token internal-token")
			},
			tokenType: authTokenTypeInternal,
			token:     "internal-token",
			ok:        true,
		},
		{
			name: "twitch prefixed header",
			setupReq: func(r *http.Request) {
				r.Header.Set("Authorization", "Twitch twitch-token")
			},
			tokenType: authTokenTypeTwitch,
			token:     "twitch-token",
			ok:        true,
		},
		{
			name: "basic auth password",
			setupReq: func(r *http.Request) {
				r.SetBasicAuth("user", "basic-token")
			},
			tokenType: authTokenTypeLegacy,
			token:     "basic-token",
			ok:        true,
		},
		{
			name: "unknown prefix",
			setupReq: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer no-thanks")
			},
			tokenType: authTokenTypeUnknown,
			token:     "",
			ok:        false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, "/", nil)
			require.NoError(t, err)
			tc.setupReq(req)

			tokenType, token, ok := getAuthorizationTokenFromRequest(req)
			require.Equal(t, tc.tokenType, tokenType)
			require.Equal(t, tc.token, token)
			require.Equal(t, tc.ok, ok)
		})
	}
}

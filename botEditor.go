package main

import (
	"fmt"
	"net/http"
	"strings"
)

func getAuthorizationFromRequest(r *http.Request) (string, error) {
	_, token, hadPrefix := strings.Cut(r.Header.Get("Authorization"), " ")
	if !hadPrefix {
		return "", fmt.Errorf("no authorization provided")
	}

	_, user, _, _, err := editorTokenService.ValidateLoginToken(token) //nolint:dogsled // Required at other places
	if err != nil {
		return "", fmt.Errorf("getting authorized user: %w", err)
	}

	if user == "" {
		user = "API-User"
	}

	return user, nil
}

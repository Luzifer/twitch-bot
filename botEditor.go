package main

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

func getAuthorizationFromRequest(r *http.Request) (string, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return "", fmt.Errorf("no authorization provided")
	}

	_, user, _, _, err := editorTokenService.ValidateLoginToken(token) //nolint:dogsled // Required at other places

	if user == "" {
		user = "API-User"
	}

	return user, errors.Wrap(err, "getting authorized user")
}

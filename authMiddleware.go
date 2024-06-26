package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/gofrs/uuid/v3"
	"github.com/pkg/errors"
	"golang.org/x/crypto/argon2"
)

const (
	// OWASP recommendations - 2023-07-07
	// https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html
	argonFmt        = "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s"
	argonHashLen    = 16
	argonMemory     = 46 * 1024
	argonSaltLength = 8
	argonThreads    = 1
	argonTime       = 1
)

func fillAuthToken(token *configAuthToken) error {
	token.Token = uuid.Must(uuid.NewV4()).String()

	salt := make([]byte, argonSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return errors.Wrap(err, "reading salt")
	}

	token.Hash = fmt.Sprintf(
		argonFmt,
		argon2.Version,
		argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(argon2.IDKey([]byte(token.Token), salt, argonTime, argonMemory, argonThreads, argonHashLen)),
	)

	return nil
}

func writeAuthMiddleware(h http.Handler, module string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pass, hasBasicAuth := r.BasicAuth()

		var token string
		switch {
		case hasBasicAuth && pass != "":
			token = pass

		case r.Header.Get("Authorization") != "":
			var (
				tokenType string
				hadPrefix bool
			)

			tokenType, token, hadPrefix = strings.Cut(r.Header.Get("Authorization"), " ")
			switch {
			case !hadPrefix:
				// Legacy: Accept `Authorization: tokenhere`
				token = tokenType

			case strings.EqualFold(tokenType, "token"):
				// This is perfect: `Authorization: Token tokenhere`

			default:
				// That was unexpected: `Authorization: Bearer tokenhere` or similar
				http.Error(w, "invalid token type", http.StatusForbidden)
				return
			}

		default:
			http.Error(w, "auth not successful", http.StatusForbidden)
			return
		}

		err := authService.ValidateTokenFor(token, module)
		if err != nil {
			http.Error(w, "auth not successful", http.StatusForbidden)
			return
		}

		h.ServeHTTP(w, r)
	})
}

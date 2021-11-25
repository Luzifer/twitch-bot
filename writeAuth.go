package main

import (
	"encoding/hex"
	"net/http"

	"github.com/gofrs/uuid/v3"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"github.com/Luzifer/go_helpers/v2/str"
)

func fillAuthToken(token *configAuthToken) error {
	token.Token = uuid.Must(uuid.NewV4()).String()

	hash, err := bcrypt.GenerateFromPassword([]byte(token.Token), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, "hashing token")
	}

	token.Hash = hex.EncodeToString(hash)

	return nil
}

func writeAuthMiddleware(h http.Handler, module string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "auth not successful", http.StatusForbidden)
			return
		}

		for _, auth := range config.AuthTokens {
			rawHash, err := hex.DecodeString(auth.Hash)
			if err != nil {
				log.WithError(err).Error("Invalid token hash found")
				continue
			}

			if bcrypt.CompareHashAndPassword(rawHash, []byte(token)) != nil {
				continue
			}

			if !str.StringInSlice(module, auth.Modules) && !str.StringInSlice("*", auth.Modules) {
				continue
			}

			h.ServeHTTP(w, r)
			return
		}

		http.Error(w, "auth not successful", http.StatusForbidden)
	})
}

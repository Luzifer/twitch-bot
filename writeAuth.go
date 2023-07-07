package main

import (
	"net/http"

	"github.com/gofrs/uuid/v3"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
)

func fillAuthToken(token *configAuthToken) error {
	token.Token = uuid.Must(uuid.NewV4()).String()

	hash, err := bcrypt.GenerateFromPassword([]byte(token.Token), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, "hashing token")
	}

	token.Hash = string(hash)

	return nil
}

func writeAuthMiddleware(h http.Handler, module string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "auth not successful", http.StatusForbidden)
			return
		}

		for _, fn := range []func() error{
			// First try to validate against internal token management
			func() error { return validateAuthToken(token, module) },
			// If not successful validate against Twitch and check for bot-editors
			func() error { return validateTwitchBotEditorAuthToken(token) },
		} {
			if err := fn(); err != nil {
				continue
			}

			h.ServeHTTP(w, r)
			return
		}

		http.Error(w, "auth not successful", http.StatusForbidden)
	})
}

func validateAuthToken(token string, modules ...string) error {
	for _, auth := range config.AuthTokens {
		if auth.validate(token) != nil {
			continue
		}

		for _, reqMod := range modules {
			if !str.StringInSlice(reqMod, auth.Modules) && !str.StringInSlice("*", auth.Modules) {
				return errors.New("missing module in auth")
			}
		}

		return nil // We found a matching token and it has all required tokens
	}

	return errors.New("no matching token")
}

func validateTwitchBotEditorAuthToken(token string) error {
	tc := twitch.New(cfg.TwitchClient, cfg.TwitchClientSecret, token, "")

	id, user, err := tc.GetAuthorizedUser()
	if err != nil {
		return errors.Wrap(err, "getting authorized user")
	}

	if !str.StringInSlice(user, config.BotEditors) && !str.StringInSlice(id, config.BotEditors) {
		return errors.New("user is not an bot-edtior")
	}

	return nil
}

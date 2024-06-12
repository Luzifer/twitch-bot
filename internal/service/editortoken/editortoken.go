// Package editortoken utilizes JWT to create / validate a token for
// the frontend
package editortoken

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/Luzifer/twitch-bot/v3/pkg/database"
	"github.com/golang-jwt/jwt/v5"
)

const (
	coreMetaSigningKey = "editortoken:signing-key"
	tokenValidity      = 24 * time.Hour
)

type (
	claims struct {
		TwitchUser twitchUser `json:"twitchUser"`

		jwt.RegisteredClaims
	}

	twitchUser struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	// Service manages the permission database
	Service struct{ db database.Connector }
)

// New creates a new Service on the given database
func New(db database.Connector) *Service {
	return &Service{db}
}

// CreateLoginToken packs user-id and user name into a JWT, signs it
// and returns the signed token
func (s Service) CreateLoginToken(id, user string) (token string, expiresAt time.Time, err error) {
	cl := claims{
		TwitchUser: twitchUser{
			ID:   id,
			Name: user,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "Twitch-Bot",
			Subject:   id,
			Audience:  []string{},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenValidity)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	tok := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, cl)

	priv, err := s.getSigningKey()
	if err != nil {
		return "", expiresAt, fmt.Errorf("getting signing key: %w", err)
	}

	if token, err = tok.SignedString(priv); err != nil {
		return "", expiresAt, fmt.Errorf("signing token: %w", err)
	}

	return token, cl.ExpiresAt.Time, nil
}

// ValidateLoginToken takes a token, validates it with the stored
// key and returns the twitch-id and the user-name from the token
func (s Service) ValidateLoginToken(token string) (id, user string, expiresAt time.Time, err error) {
	var cl claims

	tok, err := jwt.ParseWithClaims(token, &cl, func(*jwt.Token) (any, error) {
		priv, err := s.getSigningKey()
		if err != nil {
			return nil, fmt.Errorf("getting private key: %w", err)
		}

		return priv.Public(), nil
	})
	if err != nil {
		// Something went wrong when parsing & validating
		return "", "", expiresAt, fmt.Errorf("validating token: %w", err)
	}

	if claims, ok := tok.Claims.(*claims); ok {
		// We had no error and the claims are our claims
		return claims.TwitchUser.ID, claims.TwitchUser.Name, claims.ExpiresAt.Time, nil
	}

	// We had no error but were not able to convert the claims
	return "", "", expiresAt, fmt.Errorf("unknown claims type")
}

func (s Service) getSigningKey() (priv ed25519.PrivateKey, err error) {
	err = s.db.ReadEncryptedCoreMeta(coreMetaSigningKey, &priv)
	switch {
	case err == nil:
		// We read the previously generated key
		return priv, nil

	case errors.Is(err, database.ErrCoreMetaNotFound):
		// We don't have a key yet or the key was wiped for some reason,
		// we generate a new one which automatically is stored for later
		// retrieval.
		if priv, err = s.generateSigningKey(); err != nil {
			return nil, fmt.Errorf("creating signing key: %w", err)
		}

		return priv, nil

	default:
		// Something went wrong, bail.
		return nil, fmt.Errorf("reading signing key: %w", err)
	}
}

func (s Service) generateSigningKey() (ed25519.PrivateKey, error) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating key: %w", err)
	}

	if err = s.db.StoreEncryptedCoreMeta(coreMetaSigningKey, priv); err != nil {
		return nil, fmt.Errorf("storing signing key: %w", err)
	}

	return priv, nil
}

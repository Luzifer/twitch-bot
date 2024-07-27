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
	tokenValidity      = time.Hour
)

type (
	claims struct {
		Modules    []string    `json:"modules"`
		TwitchUser *twitchUser `json:"twitchUser,omitempty"`

		jwt.RegisteredClaims
	}

	twitchUser struct {
		ID   string `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
	}

	// Service manages the permission database
	Service struct{ db database.Connector }
)

// New creates a new Service on the given database
func New(db database.Connector) *Service {
	return &Service{db}
}

// CreateUserToken packs user-id and user name into a JWT, signs it
// and returns the signed token
func (s Service) CreateUserToken(id, user string, modules []string) (token string, expiresAt time.Time, err error) {
	cl := claims{
		Modules: modules,
		TwitchUser: &twitchUser{
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

	if token, err = s.createTokenFromClaims(cl); err != nil {
		return "", time.Time{}, fmt.Errorf("creating token: %w", err)
	}

	return token, cl.ExpiresAt.Time, nil
}

// CreateGenericModuleToken creates a non-user-bound token with the
// given modules. Pass in 0 validity to create a non-expiring token.
func (s Service) CreateGenericModuleToken(modules []string, validity time.Duration) (token string, err error) {
	cl := claims{
		Modules: modules,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "Twitch-Bot",
			Audience:  []string{},
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	if validity > 0 {
		cl.ExpiresAt = jwt.NewNumericDate(time.Now().Add(validity))
	}

	if token, err = s.createTokenFromClaims(cl); err != nil {
		return "", fmt.Errorf("creating token: %w", err)
	}

	return token, nil
}

// ValidateLoginToken takes a token, validates it with the stored
// key and returns the twitch-id and the user-name from the token
func (s Service) ValidateLoginToken(token string) (id, user string, expiresAt time.Time, modules []string, err error) {
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
		return "", "", expiresAt, nil, fmt.Errorf("validating token: %w", err)
	}

	if claims, ok := tok.Claims.(*claims); ok {
		if claims.ExpiresAt != nil {
			expiresAt = claims.ExpiresAt.Time
		}

		if claims.TwitchUser == nil {
			return "", "", expiresAt, claims.Modules, nil
		}
		// We had no error and the claims are our claims
		return claims.TwitchUser.ID, claims.TwitchUser.Name, expiresAt, claims.Modules, nil
	}

	// We had no error but were not able to convert the claims
	return "", "", expiresAt, nil, fmt.Errorf("unknown claims type")
}

func (s Service) createTokenFromClaims(cl claims) (token string, err error) {
	tok := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, cl)

	priv, err := s.getSigningKey()
	if err != nil {
		return "", fmt.Errorf("getting signing key: %w", err)
	}

	if token, err = tok.SignedString(priv); err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}
	return token, nil
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

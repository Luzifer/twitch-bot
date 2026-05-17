package database

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/Luzifer/twitch-bot/v3/internal/cryptkv"
)

const (
	instanceSaltCoreKVKey = "__internal:instance-salt"
	instanceSaltLength    = 16 // 16B = 128b
)

func (c connector) DecryptField(dec string) (plain string, err error) {
	if plain, err = cryptkv.Decrypt(c.instanceSalt, c.encryptionSecret, dec); err != nil {
		return "", fmt.Errorf("decrypting values: %w", err)
	}

	return plain, nil
}

func (c connector) EncryptField(enc string) (crypt string, err error) {
	if crypt, err = cryptkv.Encrypt(c.instanceSalt, c.encryptionSecret, enc); err != nil {
		return "", fmt.Errorf("encrypting values: %w", err)
	}

	return crypt, nil
}

func (c *connector) loadOrGenerateInstanceSalt() (err error) {
	var b64Salt string
	err = c.readCoreMeta(instanceSaltCoreKVKey, &b64Salt, nil)

	switch {
	case err == nil:
		// Salt was loaded
		salt, err := base64.StdEncoding.DecodeString(b64Salt)
		if err != nil {
			return fmt.Errorf("decoding salt: %w", err)
		}

		if l := len(salt); l != instanceSaltLength {
			return fmt.Errorf("instance salt has invalid length: %d != %d", l, instanceSaltLength)
		}

		c.instanceSalt = salt
		return nil

	case errors.Is(err, ErrCoreMetaNotFound):
		// We don't have a salt yet, create one
		salt := make([]byte, instanceSaltLength)
		n, err := rand.Read(salt)
		if err != nil {
			return fmt.Errorf("generating random salt: %w", err)
		}
		if n != instanceSaltLength {
			return fmt.Errorf("invalid salt length read: %d != %d", n, instanceSaltLength)
		}

		if err = c.storeCoreMeta(
			instanceSaltCoreKVKey,
			base64.StdEncoding.EncodeToString(salt),
			nil,
		); err != nil {
			return fmt.Errorf("storing salt in database: %w", err)
		}
		c.instanceSalt = salt
		return nil

	default:
		return fmt.Errorf("loading salt value: %w", err)
	}
}

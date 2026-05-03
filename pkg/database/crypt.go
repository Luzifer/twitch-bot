package database

import (
	"fmt"

	"github.com/Luzifer/go-openssl/v4"
)

func (c connector) DecryptField(dec string) (string, error) {
	dv, err := openssl.New().DecryptBytes(c.encryptionSecret, []byte(dec), openssl.PBKDF2SHA512)
	if err != nil {
		return "", fmt.Errorf("decrypting value: %w", err)
	}

	return string(dv), nil
}

func (c connector) EncryptField(enc string) (string, error) {
	ev, err := openssl.New().EncryptBytes(c.encryptionSecret, []byte(enc), openssl.PBKDF2SHA512)
	if err != nil {
		return "", fmt.Errorf("encrypting value: %w", err)
	}

	return string(ev), nil
}

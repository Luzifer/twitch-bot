package cryptkv

import (
	"fmt"

	"github.com/Luzifer/go-openssl/v4"
)

type (
	v1LegacyHandler struct{}
)

func (v1LegacyHandler) Decrypt(_ []byte, secret, ciphertext string) (plaintext string, err error) {
	dv, err := openssl.New().DecryptBytes(secret, []byte(ciphertext), openssl.PBKDF2SHA512)
	if err != nil {
		return "", fmt.Errorf("decrypting value: %w", err)
	}

	return string(dv), nil
}

func (v1LegacyHandler) Encrypt(_ []byte, _, _ string) (string, error) {
	return "", errLegacyEncryption
}

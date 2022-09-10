package database

import (
	"github.com/pkg/errors"

	"github.com/Luzifer/go-openssl/v4"
)

func (c connector) DecryptField(dec string) (string, error) {
	dv, err := openssl.New().DecryptBytes(c.encryptionSecret, []byte(dec), openssl.PBKDF2SHA512)
	return string(dv), errors.Wrap(err, "decrypting value")
}

func (c connector) EncryptField(enc string) (string, error) {
	ev, err := openssl.New().EncryptBytes(c.encryptionSecret, []byte(enc), openssl.PBKDF2SHA512)
	return string(ev), errors.Wrap(err, "encrypting value")
}

package cryptkv

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Luzifer/go-openssl/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testPlaintext = `{"access_token":"super-secret"}`
	testSecret    = "correct horse battery staple"
)

var testSalt = []byte("1234567890abcdef")

func TestDecryptErrors(t *testing.T) {
	for name, tc := range map[string]struct {
		ciphertext string
		err        string
	}{
		"handler-error": {
			ciphertext: "cryptkv2::%%%",
			err:        "decrypting secret",
		},
		"invalid-prefix": {
			ciphertext: "2::data",
			err:        "missing prefix",
		},
		"invalid-version": {
			ciphertext: "cryptkv0::data",
			err:        "invalid proto version",
		},
		"missing-prefix": {
			ciphertext: "data",
			err:        "missing prefix",
		},
		"unparseable-version": {
			ciphertext: "cryptkvx::data",
			err:        "parsing proto-version",
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Decrypt(testSalt, testSecret, tc.ciphertext)
			require.Error(t, err)
			assert.ErrorContains(t, err, tc.err)
		})
	}
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	ciphertext, err := Encrypt(testSalt, testSecret, testPlaintext)
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(ciphertext, fmt.Sprintf("cryptkv%d::", currentProto)))
	assert.NotContains(t, ciphertext, testPlaintext)

	plaintext, err := Decrypt(testSalt, testSecret, ciphertext)
	require.NoError(t, err)
	assert.JSONEq(t, testPlaintext, plaintext)
}

func TestEncryptHandlerError(t *testing.T) {
	origProto := currentProto
	currentProto = protoVersionLegacyOpenSSL
	t.Cleanup(func() { currentProto = origProto })

	_, err := Encrypt(testSalt, testSecret, testPlaintext)
	require.Error(t, err)
	assert.ErrorContains(t, err, "encrypting data")
}

func TestLegacyOpenSSLDecryptThroughDispatcher(t *testing.T) {
	ciphertext, err := openssl.New().EncryptBytes(testSecret, []byte(testPlaintext), openssl.PBKDF2SHA512)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(ciphertext), "U2FsdGVkX1"))

	plaintext, err := Decrypt(testSalt, testSecret, string(ciphertext))
	require.NoError(t, err)
	assert.JSONEq(t, testPlaintext, plaintext)
}

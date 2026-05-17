package cryptkv

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestV1EncryptDisabled(t *testing.T) {
	_, err := v1LegacyHandler{}.Encrypt(nil, "", "")
	require.ErrorIs(t, err, errLegacyEncryption)
}

func TestV1DecryptError(t *testing.T) {
	_, err := v1LegacyHandler{}.Decrypt(nil, testSecret, "invalid")
	require.Error(t, err)
	require.ErrorContains(t, err, "decrypting value")
}

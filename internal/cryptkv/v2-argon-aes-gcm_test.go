package cryptkv

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestV2DecryptErrors(t *testing.T) {
	ciphertext, err := v2ArgonAESGCM{}.Encrypt(testSalt, testSecret, testPlaintext)
	require.NoError(t, err)

	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	require.NoError(t, err)
	raw[len(raw)-1] ^= 0xff

	for name, tc := range map[string]struct {
		ciphertext string
		err        string
	}{
		"bad-authentication": {
			ciphertext: base64.StdEncoding.EncodeToString(raw),
			err:        "decrypting data",
		},
		"bad-base64": {
			ciphertext: "%%%",
			err:        "decoding base64 data",
		},
		"too-short": {
			ciphertext: base64.StdEncoding.EncodeToString([]byte("short")),
			err:        "too short",
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := v2ArgonAESGCM{}.Decrypt(testSalt, testSecret, tc.ciphertext)
			require.Error(t, err)
			assert.ErrorContains(t, err, tc.err)
		})
	}
}

func TestV2GetCipherCacheTypeMismatch(t *testing.T) {
	kdfResultCache = make(map[string]any)
	t.Cleanup(func() { kdfResultCache = make(map[string]any) })
	setKDFResult(protoVersionV2, testSalt, testSecret, 42)

	_, err := v2ArgonAESGCM{}.Encrypt(testSalt, testSecret, testPlaintext)
	require.Error(t, err)
	assert.ErrorContains(t, err, "getting cached KDF result")
}

func TestV2Roundtrip(t *testing.T) {
	ciphertext, err := v2ArgonAESGCM{}.Encrypt(testSalt, testSecret, testPlaintext)
	require.NoError(t, err)

	plaintext, err := v2ArgonAESGCM{}.Decrypt(testSalt, testSecret, ciphertext)
	require.NoError(t, err)
	assert.JSONEq(t, testPlaintext, plaintext)
}

package cryptkv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKDFCache(t *testing.T) {
	kdfResultCache = make(map[string]any)
	t.Cleanup(func() { kdfResultCache = make(map[string]any) })

	_, err := getKDFResult[string](protoVersionV2, testSalt, testSecret)
	require.ErrorIs(t, err, errKDFResultNotFound)

	setKDFResult(protoVersionV2, testSalt, testSecret, "cached")

	res, err := getKDFResult[string](protoVersionV2, testSalt, testSecret)
	require.NoError(t, err)
	assert.Equal(t, "cached", res)
}

func TestKDFCacheKey(t *testing.T) {
	baseKey := deriveKDFResultKey(protoVersionV2, testSalt, testSecret)

	assert.NotEqual(t, baseKey, deriveKDFResultKey(protoVersionLegacyOpenSSL, testSalt, testSecret))
	assert.NotEqual(t, baseKey, deriveKDFResultKey(protoVersionV2, []byte("different salt!!"), testSecret))
	assert.NotEqual(t, baseKey, deriveKDFResultKey(protoVersionV2, testSalt, "different secret"))
}

func TestKDFCacheTypeMismatch(t *testing.T) {
	kdfResultCache = make(map[string]any)
	t.Cleanup(func() { kdfResultCache = make(map[string]any) })
	setKDFResult(protoVersionV2, testSalt, testSecret, 42)

	_, err := getKDFResult[string](protoVersionV2, testSalt, testSecret)
	require.Error(t, err)
	assert.ErrorContains(t, err, "casting")
}

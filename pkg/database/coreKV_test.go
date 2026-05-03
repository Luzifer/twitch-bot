package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCoreMetaRoundtrip(t *testing.T) {
	dbc := GetTestDatabase(t)

	var (
		arbitrary struct{ A string }
		testKey   = "arbitrary"
	)

	require.ErrorIs(t, dbc.ReadCoreMeta(testKey, &arbitrary), ErrCoreMetaNotFound, "expected core_kv not to contain key after init")

	checkWriteRead := func(testString string) {
		arbitrary.A = testString
		require.NoError(t, dbc.StoreCoreMeta(testKey, arbitrary), "storing core_kv")

		arbitrary.A = "" // Clear to test unmarshal
		require.NoError(t, dbc.ReadCoreMeta(testKey, &arbitrary), "reading core_kv")

		assert.Equal(t, testString, arbitrary.A, "metadata equals")
	}

	checkWriteRead("just a string")         // Turn one: Init from not existing
	checkWriteRead("another random string") // Turn two: Overwrite
}

func TestCoreMetaEncryption(t *testing.T) {
	dbc := GetTestDatabase(t)

	var (
		arbitrary  struct{ A string }
		testKey    = "arbitrary"
		testString = "foobar"
	)

	arbitrary.A = testString
	require.NoError(t, dbc.StoreEncryptedCoreMeta(testKey, arbitrary), "storing encrypted core meta")

	require.Error(t, dbc.ReadCoreMeta(testKey, &arbitrary), "reading encrypted meta without decryption succeeded")

	arbitrary.A = ""

	require.NoError(t, dbc.ReadEncryptedCoreMeta(testKey, &arbitrary), "reading encrypted meta")
	assert.Equal(t, testString, arbitrary.A, "unexpected value")
}

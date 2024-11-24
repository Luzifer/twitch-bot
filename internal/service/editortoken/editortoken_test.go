package editortoken

import (
	"crypto/ed25519"
	"testing"
	"time"

	"github.com/Luzifer/twitch-bot/v3/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateToken(t *testing.T) {
	dbc := database.GetTestDatabase(t)
	s := New(dbc)

	// Fresh database, no key stored, the key should be generated and
	// stored
	pk1, err := s.getSigningKey()
	require.NoError(t, err)
	assert.IsType(t, ed25519.PrivateKey{}, pk1)

	// Now database should contain key
	var dbpk ed25519.PrivateKey
	err = dbc.ReadCoreMeta(coreMetaSigningKey, &dbpk)
	require.Error(t, err, "Key must not be readable with plain func")
	err = dbc.ReadEncryptedCoreMeta(coreMetaSigningKey, &dbpk)
	require.NoError(t, err)

	// When fetching the key again it should be the same as before
	pk2, err := s.getSigningKey()
	require.NoError(t, err)
	assert.Equal(t, pk1, pk2)
	assert.Equal(t, dbpk, pk2)
}

func TestTokenFlow(t *testing.T) {
	dbc := database.GetTestDatabase(t)
	s := New(dbc)

	var (
		id   = "123456"
		user = "example"
	)

	tok, expiresAt, err := s.CreateUserToken(id, user, []string{"*"})
	require.NoError(t, err)
	assert.True(t, expiresAt.After(time.Now().Add(tokenValidity-time.Minute)))

	tid, tuser, texpiresAt, modules, err := s.ValidateLoginToken(tok)
	require.NoError(t, err)
	assert.Equal(t, id, tid)
	assert.Equal(t, user, tuser)
	assert.Equal(t, expiresAt, texpiresAt)
	assert.Equal(t, []string{"*"}, modules)

	// Generic without expiry
	tok, err = s.CreateGenericModuleToken([]string{"test"}, 0)
	require.NoError(t, err)

	tid, tuser, texpiresAt, modules, err = s.ValidateLoginToken(tok)
	require.NoError(t, err)
	assert.Equal(t, "", tid)
	assert.Equal(t, "", tuser)
	assert.Equal(t, time.Time{}, texpiresAt)
	assert.Equal(t, []string{"test"}, modules)

	// Generic with expiry
	tok, err = s.CreateGenericModuleToken([]string{"test"}, time.Minute)
	require.NoError(t, err)

	tid, tuser, texpiresAt, modules, err = s.ValidateLoginToken(tok)
	require.NoError(t, err)
	assert.Equal(t, "", tid)
	assert.Equal(t, "", tuser)
	assert.True(t, time.Now().Add(time.Minute+time.Second).After(texpiresAt))
	assert.Equal(t, []string{"test"}, modules)
}

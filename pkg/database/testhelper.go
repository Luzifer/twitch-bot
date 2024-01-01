package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// GetTestDatabase returns a Connector to an in-mem SQLite DB
func GetTestDatabase(t *testing.T) Connector {
	dbc, err := New("sqlite", "file::memory:?cache=shared", "encpass")
	require.NoError(t, err, "creating database connector")
	t.Cleanup(func() {
		if err := dbc.Close(); err != nil {
			t.Logf("closing in-mem database: %s", err)
		}
	})

	return dbc
}

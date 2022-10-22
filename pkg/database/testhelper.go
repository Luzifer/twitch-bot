package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func GetTestDatabase(t *testing.T) Connector {
	dbc, err := New("sqlite", "file::memory:?cache=shared", "encpass")
	require.NoError(t, err, "creating database connector")
	t.Cleanup(func() { dbc.Close() })

	return dbc
}

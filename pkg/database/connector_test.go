package database

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testEncryptionPass = "password123"

func TestNewConnector(t *testing.T) {
	cStrings := map[string]string{
		"filesystem": path.Join(t.TempDir(), "storage.db"),
		"memory":     "file::memory:?cache=shared",
	}

	for name := range cStrings {
		t.Run(name, func(t *testing.T) {
			dbc, err := New("sqlite", cStrings[name], testEncryptionPass)
			require.NoError(t, err, "creating database connector")
			t.Cleanup(func() { dbc.Close() })

			row := dbc.DB().Raw("SELECT count(1) AS tables FROM sqlite_master WHERE type='table' AND name='core_kvs';")

			var count int
			assert.NoError(t, row.Scan(&count).Error, "reading table count result")

			assert.Equal(t, 1, count)
		})
	}
}

func TestPatchSQLiteConnString(t *testing.T) {
	for in, out := range map[string]string{
		"storage.db":                 "storage.db?_pragma=locking_mode(EXCLUSIVE)&_pragma=synchronous(FULL)",
		"file::memory:?cache=shared": "file::memory:?_pragma=locking_mode(EXCLUSIVE)&_pragma=synchronous(FULL)&cache=shared",
	} {
		cs, err := patchSQLiteConnString(in)
		require.NoError(t, err, "patching conn string %q", in)
		assert.Equal(t, out, cs, "patching conn string %q", in)
	}
}

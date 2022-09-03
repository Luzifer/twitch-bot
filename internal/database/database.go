// Package database represents a connector to the sqlite storage
// backend to store persistent data from core and plugins
package database

import (
	"io/fs"

	"github.com/jmoiron/sqlx"

	// Included support for pure-go sqlite
	_ "github.com/glebarez/go-sqlite"
)

type (
	// Connector represents a database connection having some extra
	// convenience methods
	Connector interface {
		Close() error
		DB() *sqlx.DB
		Migrate(module string, migrations MigrationStorage) error
		ReadCoreMeta(key string, value any) error
		StoreCoreMeta(key string, value any) error
	}

	// MigrationStorage represents a file storage containing migration
	// files to migrate a namespace to its desired state. The files
	// MUST be named in the schema `[0-9]+\.sql`.
	//
	// The storage is scanned recursively and all files are then
	// string-sorted by their base-name (`/migrations/001.sql => 001.sql`).
	// The last executed number is stored in numeric format, the next
	// migration which basename evaluates to higher numeric will be
	// executed.
	//
	// Numbers MUST be consecutive and MUST NOT leave out a number. A
	// missing number will result in the migration processing not to
	// catch up any migration afterwards.
	//
	// The first migration MUST be number 1
	//
	// Previously executed migrations MUST NOT be modified!
	MigrationStorage interface {
		ReadDir(name string) ([]fs.DirEntry, error)
		ReadFile(name string) ([]byte, error)
	}
)

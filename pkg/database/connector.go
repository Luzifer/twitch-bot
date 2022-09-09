package database

import (
	"embed"
	"regexp"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type (
	connector struct {
		db               *sqlx.DB
		encryptionSecret string
	}
)

var (
	// ErrCoreMetaNotFound is the error thrown when reading a non-existent
	// core_kv key
	ErrCoreMetaNotFound = errors.New("core meta entry not found")

	//go:embed schema/**
	schema embed.FS

	migrationFilename = regexp.MustCompile(`^([0-9]+)\.sql$`)
)

// New creates a new Connector with the given driver and database
func New(driverName, dataSourceName, encryptionSecret string) (Connector, error) {
	db, err := sqlx.Connect(driverName, dataSourceName)
	if err != nil {
		return nil, errors.Wrap(err, "connecting database")
	}

	db.SetConnMaxIdleTime(0)
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(1)

	conn := &connector{
		db:               db,
		encryptionSecret: encryptionSecret,
	}
	return conn, errors.Wrap(conn.applyCoreSchema(), "applying core schema")
}

func (c connector) Close() error {
	return errors.Wrap(c.db.Close(), "closing database")
}

func (c connector) DB() *sqlx.DB {
	return c.db
}

func (c connector) applyCoreSchema() error {
	coreSQL, err := schema.ReadFile("schema/core.sql")
	if err != nil {
		return errors.Wrap(err, "reading core.sql content")
	}

	if _, err = c.db.Exec(string(coreSQL)); err != nil {
		return errors.Wrap(err, "applying core schema")
	}

	return errors.Wrap(c.Migrate("core", NewEmbedFSMigrator(schema, "schema")), "applying core migration")
}

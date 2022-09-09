package database

import (
	"bytes"
	"database/sql"
	"embed"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type (
	connector struct {
		db *sqlx.DB
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
func New(driverName, dataSourceName string) (Connector, error) {
	db, err := sqlx.Connect(driverName, dataSourceName)
	if err != nil {
		return nil, errors.Wrap(err, "connecting database")
	}

	db.SetConnMaxIdleTime(0)
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(1)

	conn := &connector{db: db}
	return conn, errors.Wrap(conn.applyCoreSchema(), "applying core schema")
}

func (c connector) Close() error {
	return errors.Wrap(c.db.Close(), "closing database")
}

func (c connector) DB() *sqlx.DB {
	return c.db
}

// ReadCoreMeta reads an entry of the core_kv table specified by
// the given `key` and unmarshals it into the `value`. The value must
// be a valid variable to `json.NewDecoder(...).Decode(value)`
// (pointer to struct, string, int, ...). In case the key does not
// exist a check to 'errors.Is(err, sql.ErrNoRows)' will succeed
func (c connector) ReadCoreMeta(key string, value any) error {
	var data struct{ Key, Value string }
	data.Key = key

	if err := c.db.Get(&data, "SELECT * FROM core_kv WHERE key = $1", data.Key); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCoreMetaNotFound
		}
		return errors.Wrap(err, "querying core meta table")
	}

	if data.Value == "" {
		return errors.New("empty value returned")
	}

	if err := json.NewDecoder(strings.NewReader(data.Value)).Decode(value); err != nil {
		return errors.Wrap(err, "JSON decoding value")
	}

	return nil
}

// StoreCoreMeta stores an entry to the core_kv table soecified by
// the given `key`. The value given must be a valid variable to
// `json.NewEncoder(...).Encode(value)`.
func (c connector) StoreCoreMeta(key string, value any) error {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(value); err != nil {
		return errors.Wrap(err, "JSON encoding value")
	}

	_, err := c.db.NamedExec(
		"INSERT INTO core_kv (key, value) VALUES (:key, :value) ON CONFLICT DO UPDATE SET value=excluded.value;",
		map[string]any{
			"key":   key,
			"value": buf.String(),
		},
	)

	return errors.Wrap(err, "upserting core meta value")
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

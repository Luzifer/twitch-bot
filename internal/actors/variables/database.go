package variables

import (
	"database/sql"
	"embed"

	"github.com/pkg/errors"
)

//go:embed schema/**
var schema embed.FS

func getVariable(key string) (string, error) {
	row := db.DB().QueryRow(
		`SELECT value
			FROM variables
			WHERE name = $1`,
		key,
	)

	var vc string
	err := row.Scan(&vc)
	switch {
	case err == nil:
		return vc, nil

	case errors.Is(err, sql.ErrNoRows):
		return "", nil // Compatibility to old behavior

	default:
		return "", errors.Wrap(err, "getting value from database")
	}
}

func setVariable(key, value string) error {
	_, err := db.DB().Exec(
		`INSERT INTO variables
			(name, value)
			VALUES ($1, $2)
			ON CONFLICT DO UPDATE
				SET value = excluded.value;`,
		key, value,
	)

	return errors.Wrap(err, "updating value in database")
}

func removeVariable(key string) error {
	_, err := db.DB().Exec(
		`DELETE FROM variables
			WHERE name = $1;`,
		key,
	)

	return errors.Wrap(err, "deleting value in database")
}

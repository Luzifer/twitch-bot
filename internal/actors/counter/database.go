package counter

import (
	"database/sql"
	"embed"

	"github.com/pkg/errors"
)

//go:embed schema/**
var schema embed.FS

func getCounterValue(counter string) (int64, error) {
	row := db.DB().QueryRow(
		`SELECT value
			FROM counters
			WHERE name = $1`,
		counter,
	)

	var cv int64
	err := row.Scan(&cv)
	switch {
	case err == nil:
		return cv, nil

	case errors.Is(err, sql.ErrNoRows):
		return 0, nil

	default:
		return 0, errors.Wrap(err, "querying counter")
	}
}

func updateCounter(counter string, value int64, absolute bool) error {
	if !absolute {
		cv, err := getCounterValue(counter)
		if err != nil {
			return errors.Wrap(err, "getting previous value")
		}

		value += cv
	}

	_, err := db.DB().Exec(
		`INSERT INTO counters
			(name, value)
			VALUES ($1, $2)
			ON CONFLICT DO UPDATE
				SET value = excluded.value;`,
		counter, value,
	)

	return errors.Wrap(err, "storing counter value")
}

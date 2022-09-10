package punish

import (
	"database/sql"
	"embed"
	"strings"
	"time"

	"github.com/pkg/errors"
)

//go:embed schema/**
var schema embed.FS

func calculateCurrentPunishments() error {
	rows, err := db.DB().Query(
		`SELECT key, last_level, executed, cooldown
			FROM punish_levels;`,
	)
	if err != nil {
		return errors.Wrap(err, "querying punish_levels")
	}

	for rows.Next() {
		if err = rows.Err(); err != nil {
			return errors.Wrap(err, "advancing rows")
		}

		var (
			key                           string
			lastLevel, executed, cooldown int64

			actUpdate bool
		)
		if err = rows.Scan(&key, &lastLevel, &executed, &cooldown); err != nil {
			return errors.Wrap(err, "advancing rows")
		}

		lvl := &levelConfig{
			LastLevel: int(lastLevel),
			Cooldown:  time.Duration(cooldown),
			Executed:  time.Unix(executed, 0),
		}

		for {
			cooldownTime := lvl.Executed.Add(lvl.Cooldown)
			if cooldownTime.After(time.Now()) {
				break
			}

			lvl.Executed = cooldownTime
			lvl.LastLevel--
			actUpdate = true
		}

		// Level 0 is the first punishment level, so only remove if it drops below 0
		if lvl.LastLevel < 0 {
			if err = deletePunishmentForKey(key); err != nil {
				return errors.Wrap(err, "cleaning up expired punishment")
			}
			continue
		}

		if actUpdate {
			if err = setPunishmentForKey(key, lvl); err != nil {
				return errors.Wrap(err, "updating punishment")
			}
		}
	}

	return errors.Wrap(rows.Err(), "finishing rows processing")
}

func deletePunishment(channel, user, uuid string) error {
	return deletePunishmentForKey(getDBKey(channel, user, uuid))
}

func deletePunishmentForKey(key string) error {
	_, err := db.DB().Exec(
		`DELETE FROM punish_levels
			WHERE key = $1;`,
		key,
	)

	return errors.Wrap(err, "deleting punishment info")
}

func getPunishment(channel, user, uuid string) (*levelConfig, error) {
	if err := calculateCurrentPunishments(); err != nil {
		return nil, errors.Wrap(err, "updating punishment states")
	}

	row := db.DB().QueryRow(
		`SELECT last_level, executed, cooldown
			FROM punish_levels
			WHERE key = $1;`,
		getDBKey(channel, user, uuid),
	)

	lc := &levelConfig{LastLevel: -1}

	var lastLevel, executed, cooldown int64
	err := row.Scan(&lastLevel, &executed, &cooldown)
	switch {
	case err == nil:
		lc.LastLevel = int(lastLevel)
		lc.Cooldown = time.Duration(cooldown)
		lc.Executed = time.Unix(executed, 0)

		return lc, nil

	case errors.Is(err, sql.ErrNoRows):
		return lc, nil

	default:
		return nil, errors.Wrap(err, "getting punishment from database")
	}
}

func setPunishment(channel, user, uuid string, lc *levelConfig) error {
	return setPunishmentForKey(getDBKey(channel, user, uuid), lc)
}

func setPunishmentForKey(key string, lc *levelConfig) error {
	_, err := db.DB().Exec(
		`INSERT INTO punish_levels
			(key, last_level, executed, cooldown)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT DO UPDATE
				SET last_level = excluded.last_level,
					executed = excluded.executed,
					cooldown = excluded.cooldown;`,
		key,
		lc.LastLevel, lc.Executed.UTC().Unix(), int64(lc.Cooldown),
	)

	return errors.Wrap(err, "updating punishment info")
}

func getDBKey(channel, user, uuid string) string {
	return strings.Join([]string{channel, user, uuid}, "::")
}

package v2migrator

import (
	"time"

	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/pkg/database"
)

type (
	storageModPunish struct {
		ActiveLevels map[string]*struct {
			LastLevel int           `json:"last_level"`
			Executed  time.Time     `json:"executed"`
			Cooldown  time.Duration `json:"cooldown"`
		} `json:"active_levels"`
	}
)

func (s storageModPunish) migrate(db database.Connector) (err error) {
	for key, lc := range s.ActiveLevels {
		if _, err = db.DB().Exec(
			`INSERT INTO punish_levels
				(key, last_level, executed, cooldown)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT DO UPDATE
					SET last_level = excluded.last_level,
						executed = excluded.executed,
						cooldown = excluded.cooldown;`,
			key,
			lc.LastLevel, lc.Executed.UTC().Unix(), int64(lc.Cooldown),
		); err != nil {
			return errors.Wrap(err, "updating punishment info")
		}
	}

	return nil
}

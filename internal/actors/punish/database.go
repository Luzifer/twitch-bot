package punish

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Luzifer/go_helpers/v2/backoff"
	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/database"
)

type (
	punishLevel struct {
		Key string `gorm:"primaryKey"`

		LastLevel int
		Executed  time.Time
		Cooldown  time.Duration
	}
)

func calculateCurrentPunishments(db database.Connector) (err error) {
	var ps []punishLevel
	if err = helpers.Retry(func() error { return db.DB().Find(&ps).Error }); err != nil {
		return errors.Wrap(err, "querying punish_levels")
	}

	for _, p := range ps {
		var (
			actUpdate bool
			lvl       = &levelConfig{
				LastLevel: p.LastLevel,
				Executed:  p.Executed,
				Cooldown:  p.Cooldown,
			}
		)

		for {
			cooldownTime := lvl.Executed.Add(lvl.Cooldown)
			if cooldownTime.After(time.Now().UTC()) {
				break
			}

			lvl.Executed = cooldownTime
			lvl.LastLevel--
			actUpdate = true
		}

		// Level 0 is the first punishment level, so only remove if it drops below 0
		if lvl.LastLevel < 0 {
			if err = deletePunishmentForKey(db, p.Key); err != nil {
				return errors.Wrap(err, "cleaning up expired punishment")
			}
			continue
		}

		if actUpdate {
			if err = setPunishmentForKey(db, p.Key, lvl); err != nil {
				return errors.Wrap(err, "updating punishment")
			}
		}
	}

	return nil
}

func deletePunishment(db database.Connector, channel, user, uuid string) error {
	return deletePunishmentForKey(db, getDBKey(channel, user, uuid))
}

func deletePunishmentForKey(db database.Connector, key string) error {
	return errors.Wrap(
		helpers.RetryTransaction(db.DB(), func(tx *gorm.DB) error {
			return tx.Delete(&punishLevel{}, "key = ?", key).Error
		}),
		"deleting punishment info",
	)
}

func getPunishment(db database.Connector, channel, user, uuid string) (*levelConfig, error) {
	if err := calculateCurrentPunishments(db); err != nil {
		return nil, errors.Wrap(err, "updating punishment states")
	}

	var (
		lc = &levelConfig{LastLevel: -1}
		p  punishLevel
	)

	err := helpers.Retry(func() error {
		err := db.DB().First(&p, "key = ?", getDBKey(channel, user, uuid)).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return backoff.NewErrCannotRetry(err)
		}
		return err
	})
	switch {
	case err == nil:
		return &levelConfig{
			LastLevel: p.LastLevel,
			Executed:  p.Executed,
			Cooldown:  p.Cooldown,
		}, nil

	case errors.Is(err, gorm.ErrRecordNotFound):
		return lc, nil

	default:
		return nil, errors.Wrap(err, "getting punishment from database")
	}
}

func setPunishment(db database.Connector, channel, user, uuid string, lc *levelConfig) error {
	return setPunishmentForKey(db, getDBKey(channel, user, uuid), lc)
}

func setPunishmentForKey(db database.Connector, key string, lc *levelConfig) error {
	if lc == nil {
		return errors.New("nil levelConfig given")
	}

	return errors.Wrap(
		helpers.RetryTransaction(db.DB(), func(tx *gorm.DB) error {
			return tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				UpdateAll: true,
			}).Create(punishLevel{
				Key:       key,
				LastLevel: lc.LastLevel,
				Executed:  lc.Executed,
				Cooldown:  lc.Cooldown,
			}).Error
		}),
		"updating punishment info",
	)
}

func getDBKey(channel, user, uuid string) string {
	return strings.Join([]string{channel, user, uuid}, "::")
}

package counter

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/database"
)

type (
	Counter struct {
		Name  string `gorm:"primaryKey"`
		Value int64
	}
)

func GetCounterValue(db database.Connector, counterName string) (int64, error) {
	var c Counter

	err := helpers.Retry(func() error {
		err := db.DB().First(&c, "name = ?", counterName).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}

		return err
	})

	return c.Value, errors.Wrap(err, "querying counter")
}

func UpdateCounter(db database.Connector, counterName string, value int64, absolute bool) error {
	if !absolute {
		cv, err := GetCounterValue(db, counterName)
		if err != nil {
			return errors.Wrap(err, "getting previous value")
		}

		value += cv
	}

	return errors.Wrap(
		helpers.RetryTransaction(db.DB(), func(tx *gorm.DB) error {
			return tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "name"}},
				DoUpdates: clause.AssignmentColumns([]string{"value"}),
			}).Create(Counter{Name: counterName, Value: value}).Error
		}),
		"storing counter value",
	)
}

func getCounterRank(db database.Connector, prefix, name string) (rank, count int64, err error) {
	var cc []Counter

	if err = helpers.Retry(func() error {
		return db.DB().
			Order("value DESC").
			Find(&cc, "name LIKE ?", prefix+"%").
			Error
	}); err != nil {
		return 0, 0, errors.Wrap(err, "querying counters")
	}

	for i, c := range cc {
		count++
		if c.Name == name {
			rank = int64(i + 1)
		}
	}

	return rank, count, nil
}

func getCounterTopList(db database.Connector, prefix string, n int) ([]Counter, error) {
	var cc []Counter

	err := helpers.Retry(func() error {
		return db.DB().
			Order("value DESC").
			Limit(n).
			Find(&cc, "name LIKE ?", prefix+"%").
			Error
	})

	return cc, errors.Wrap(err, "querying counters")
}

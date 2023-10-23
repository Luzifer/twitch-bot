package counter

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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

	err := db.DB().First(&c, "name = ?", counterName).Error
	switch {
	case err == nil:
		return c.Value, nil

	case errors.Is(err, gorm.ErrRecordNotFound):
		return 0, nil

	default:
		return 0, errors.Wrap(err, "querying counter")
	}
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
		db.DB().Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{"value"}),
		}).Create(Counter{Name: counterName, Value: value}).Error,
		"storing counter value",
	)
}

func getCounterRank(db database.Connector, prefix, name string) (rank, count int64, err error) {
	var cc []Counter

	err = db.DB().
		Order("value DESC").
		Find(&cc, "name LIKE ?", prefix+"%").
		Error
	if err != nil {
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

	err := db.DB().
		Order("value DESC").
		Limit(n).
		Find(&cc, "name LIKE ?", prefix+"%").
		Error

	return cc, errors.Wrap(err, "querying counters")
}

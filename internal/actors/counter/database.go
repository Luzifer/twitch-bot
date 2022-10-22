package counter

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Luzifer/twitch-bot/pkg/database"
)

type (
	counter struct {
		Name  string `gorm:"primaryKey"`
		Value int64
	}
)

func GetCounterValue(db database.Connector, counterName string) (int64, error) {
	var c counter

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
		}).Create(counter{Name: counterName, Value: value}).Error,
		"storing counter value",
	)
}

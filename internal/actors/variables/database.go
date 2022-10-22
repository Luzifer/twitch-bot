package variables

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Luzifer/twitch-bot/pkg/database"
)

type (
	variable struct {
		Name  string `gorm:"primaryKey"`
		Value string
	}
)

func GetVariable(db database.Connector, key string) (string, error) {
	var v variable
	err := db.DB().First(&v, "name = ?", key).Error
	switch {
	case err == nil:
		return v.Value, nil

	case errors.Is(err, gorm.ErrRecordNotFound):
		return "", nil // Compatibility to old behavior

	default:
		return "", errors.Wrap(err, "getting value from database")
	}
}

func SetVariable(db database.Connector, key, value string) error {
	return errors.Wrap(
		db.DB().Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{"value"}),
		}).Create(variable{Name: key, Value: value}).Error,
		"updating value in database",
	)
}

func RemoveVariable(db database.Connector, key string) error {
	return errors.Wrap(
		db.DB().Delete(&variable{}, "name = ?", key).Error,
		"deleting value in database",
	)
}

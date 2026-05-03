package variables

import (
	"errors"
	"fmt"

	"github.com/Luzifer/go_helpers/backoff"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/database"
)

type (
	variable struct {
		Name  string `gorm:"primaryKey"`
		Value string
	}
)

func getVariable(db database.Connector, key string) (string, error) {
	var v variable
	err := helpers.Retry(func() error {
		err := db.DB().First(&v, "name = ?", key).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return backoff.NewErrCannotRetry(err)
		}
		return err
	})
	switch {
	case err == nil:
		return v.Value, nil

	case errors.Is(err, gorm.ErrRecordNotFound):
		return "", nil // Compatibility to old behavior

	default:
		return "", fmt.Errorf("getting value from database: %w", err)
	}
}

func setVariable(db database.Connector, key, value string) error {
	if err := helpers.RetryTransaction(db.DB(), func(tx *gorm.DB) error {
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{"value"}),
		}).Create(variable{Name: key, Value: value}).Error
	}); err != nil {
		return fmt.Errorf("updating value in database: %w", err)
	}

	return nil
}

func removeVariable(db database.Connector, key string) error {
	if err := helpers.RetryTransaction(db.DB(), func(tx *gorm.DB) error {
		return tx.Delete(&variable{}, "name = ?", key).Error
	}); err != nil {
		return fmt.Errorf("deleting value in database: %w", err)
	}

	return nil
}

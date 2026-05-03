// Package timer contains a service to store and manage timers in a database
package timer

import (
	"errors"
	"fmt"
	"time"

	"github.com/Luzifer/go_helpers/backoff"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/database"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

type (
	// Service implements a timer service
	Service struct {
		db            database.Connector
		permitTimeout time.Duration
	}

	timer struct {
		ID        string `gorm:"primaryKey"`
		ExpiresAt time.Time
	}
)

var _ plugins.TimerStore = (*Service)(nil)

// New creates a new Service
func New(db database.Connector, cronService *cron.Cron) (*Service, error) {
	s := &Service{
		db: db,
	}

	if cronService != nil {
		if _, err := cronService.AddFunc("@every 5m", s.cleanupTimers); err != nil {
			return nil, fmt.Errorf("registering timer cleanup cron: %w", err)
		}
	}

	if err := s.db.DB().AutoMigrate(&timer{}); err != nil {
		return nil, fmt.Errorf("applying migrations: %w", err)
	}

	return s, nil
}

// CopyDatabase enables the service to migrate to a new database
func (*Service) CopyDatabase(src, target *gorm.DB) error {
	return database.CopyObjects(src, target, &timer{}) //nolint:wrapcheck // Helper in own package
}

// HasTimer checks whether a timer with given ID is present
func (s Service) HasTimer(id string) (bool, error) {
	var t timer
	err := helpers.Retry(func() error {
		err := s.db.DB().First(&t, "id = ? AND expires_at >= ?", id, time.Now().UTC()).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return backoff.NewErrCannotRetry(err)
		}
		return err
	})
	switch {
	case err == nil:
		return true, nil

	case errors.Is(err, gorm.ErrRecordNotFound):
		return false, nil

	default:
		return false, fmt.Errorf("getting timer information: %w", err)
	}
}

// SetTimer sets a timer with given ID and expiry
func (s Service) SetTimer(id string, expiry time.Time) error {
	if err := helpers.RetryTransaction(s.db.DB(), func(tx *gorm.DB) error {
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"expires_at"}),
		}).Create(timer{
			ID:        id,
			ExpiresAt: expiry.UTC(),
		}).Error
	}); err != nil {
		return fmt.Errorf("storing counter in database: %w", err)
	}

	return nil
}

// UpdatePermitTimeout sets a new permit timeout for future permits
func (s *Service) UpdatePermitTimeout(d time.Duration) {
	s.permitTimeout = d
}

func (s Service) cleanupTimers() {
	if err := helpers.RetryTransaction(s.db.DB(), func(tx *gorm.DB) error {
		return tx.Delete(&timer{}, "expires_at < ?", time.Now().UTC()).Error
	}); err != nil {
		logrus.WithError(err).Error("cleaning up expired timers")
	}
}

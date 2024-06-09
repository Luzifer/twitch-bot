// Package timer contains a service to store and manage timers in a database
package timer

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Luzifer/go_helpers/v2/backoff"
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
			return nil, errors.Wrap(err, "registering timer cleanup cron")
		}
	}

	return s, errors.Wrap(s.db.DB().AutoMigrate(&timer{}), "applying migrations")
}

// CopyDatabase enables the service to migrate to a new database
func (*Service) CopyDatabase(src, target *gorm.DB) error {
	return database.CopyObjects(src, target, &timer{}) //nolint:wrapcheck // Helper in own package
}

// UpdatePermitTimeout sets a new permit timeout for future permits
func (s *Service) UpdatePermitTimeout(d time.Duration) {
	s.permitTimeout = d
}

// Cooldown timer

// AddCooldown adds a new cooldown timer
func (s Service) AddCooldown(tt plugins.TimerType, limiter, ruleID string, expiry time.Time) error {
	return s.SetTimer(s.getCooldownTimerKey(tt, limiter, ruleID), expiry)
}

// InCooldown checks whether the cooldown has expired
func (s Service) InCooldown(tt plugins.TimerType, limiter, ruleID string) (bool, error) {
	return s.HasTimer(s.getCooldownTimerKey(tt, limiter, ruleID))
}

func (Service) getCooldownTimerKey(tt plugins.TimerType, limiter, ruleID string) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(fmt.Sprintf("%d:%s:%s", tt, limiter, ruleID))))
}

// Permit timer

// AddPermit adds a new permit timer
func (s Service) AddPermit(channel, username string) error {
	return s.SetTimer(s.getPermitTimerKey(channel, username), time.Now().Add(s.permitTimeout))
}

// HasPermit checks whether a valid permit is present
func (s Service) HasPermit(channel, username string) (bool, error) {
	return s.HasTimer(s.getPermitTimerKey(channel, username))
}

func (Service) getPermitTimerKey(channel, username string) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(fmt.Sprintf(
		"%d:%s:%s",
		plugins.TimerTypePermit, channel, strings.ToLower(strings.TrimLeft(username, "@")),
	))))
}

// Generic timer

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
		return false, errors.Wrap(err, "getting timer information")
	}
}

// SetTimer sets a timer with given ID and expiry
func (s Service) SetTimer(id string, expiry time.Time) error {
	return errors.Wrap(
		helpers.RetryTransaction(s.db.DB(), func(tx *gorm.DB) error {
			return tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"expires_at"}),
			}).Create(timer{
				ID:        id,
				ExpiresAt: expiry.UTC(),
			}).Error
		}),
		"storing counter in database",
	)
}

func (s Service) cleanupTimers() {
	if err := helpers.RetryTransaction(s.db.DB(), func(tx *gorm.DB) error {
		return tx.Delete(&timer{}, "expires_at < ?", time.Now().UTC()).Error
	}); err != nil {
		logrus.WithError(err).Error("cleaning up expired timers")
	}
}

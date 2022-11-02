package timer

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Luzifer/twitch-bot/v3/pkg/database"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

type (
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

func New(db database.Connector) (*Service, error) {
	s := &Service{
		db: db,
	}

	return s, errors.Wrap(s.db.DB().AutoMigrate(&timer{}), "applying migrations")
}

func (s *Service) UpdatePermitTimeout(d time.Duration) {
	s.permitTimeout = d
}

// Cooldown timer

func (s Service) AddCooldown(tt plugins.TimerType, limiter, ruleID string, expiry time.Time) error {
	return s.SetTimer(s.getCooldownTimerKey(tt, limiter, ruleID), expiry)
}

func (s Service) InCooldown(tt plugins.TimerType, limiter, ruleID string) (bool, error) {
	return s.HasTimer(s.getCooldownTimerKey(tt, limiter, ruleID))
}

func (Service) getCooldownTimerKey(tt plugins.TimerType, limiter, ruleID string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%d:%s:%s", tt, limiter, ruleID)
	return fmt.Sprintf("sha256:%x", h.Sum(nil))
}

// Permit timer

func (s Service) AddPermit(channel, username string) error {
	return s.SetTimer(s.getPermitTimerKey(channel, username), time.Now().Add(s.permitTimeout))
}

func (s Service) HasPermit(channel, username string) (bool, error) {
	return s.HasTimer(s.getPermitTimerKey(channel, username))
}

func (Service) getPermitTimerKey(channel, username string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%d:%s:%s", plugins.TimerTypePermit, channel, strings.ToLower(strings.TrimLeft(username, "@")))
	return fmt.Sprintf("sha256:%x", h.Sum(nil))
}

// Generic timer

func (s Service) HasTimer(id string) (bool, error) {
	var t timer
	err := s.db.DB().First(&t, "id = ? AND expires_at >= ?", id, time.Now().UTC()).Error
	switch {
	case err == nil:
		return true, nil

	case errors.Is(err, gorm.ErrRecordNotFound):
		return false, nil

	default:
		return false, errors.Wrap(err, "getting timer information")
	}
}

func (s Service) SetTimer(id string, expiry time.Time) error {
	return errors.Wrap(
		s.db.DB().Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"expires_at"}),
		}).Create(timer{
			ID:        id,
			ExpiresAt: expiry.UTC(),
		}).Error,
		"storing counter in database",
	)
}

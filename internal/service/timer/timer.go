package timer

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/internal/database"
	"github.com/Luzifer/twitch-bot/plugins"
)

type (
	Service struct {
		db            database.Connector
		permitTimeout time.Duration
	}
)

var _ plugins.TimerStore = (*Service)(nil)

func New(db database.Connector) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) UpdatePermitTimeout(d time.Duration) {
	s.permitTimeout = d
}

// Cooldown timer

func (s Service) AddCooldown(tt plugins.TimerType, limiter, ruleID string, expiry time.Time) error {
	return s.setTimer(plugins.TimerTypeCooldown, s.getCooldownTimerKey(tt, limiter, ruleID), expiry)
}

func (s Service) InCooldown(tt plugins.TimerType, limiter, ruleID string) (bool, error) {
	return s.hasTimer(s.getCooldownTimerKey(tt, limiter, ruleID))
}

func (Service) getCooldownTimerKey(tt plugins.TimerType, limiter, ruleID string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%d:%s:%s", tt, limiter, ruleID)
	return fmt.Sprintf("sha256:%x", h.Sum(nil))
}

// Permit timer

func (s Service) AddPermit(channel, username string) error {
	return s.setTimer(plugins.TimerTypePermit, s.getPermitTimerKey(channel, username), time.Now().Add(s.permitTimeout))
}

func (s Service) HasPermit(channel, username string) (bool, error) {
	return s.hasTimer(s.getPermitTimerKey(channel, username))
}

func (Service) getPermitTimerKey(channel, username string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%d:%s:%s", plugins.TimerTypePermit, channel, strings.ToLower(strings.TrimLeft(username, "@")))
	return fmt.Sprintf("sha256:%x", h.Sum(nil))
}

// Generic timer

func (s Service) hasTimer(id string) (bool, error) {
	row := s.db.DB().QueryRow(
		`SELECT COUNT(1) as active_counters
			FROM timers
			WHERE id = $1 AND expires_at >= $2`,
		id, time.Now().UTC().Unix(),
	)

	var nCounters int64
	if err := row.Scan(&nCounters); err != nil {
		return false, errors.Wrap(err, "getting active counters from database")
	}

	return nCounters > 0, nil
}

func (s Service) setTimer(kind plugins.TimerType, id string, expiry time.Time) error {
	_, err := s.db.DB().Exec(
		`INSERT INTO timers
			(id, expires_at)
			VALUES ($1, $2)
			ON CONFLICT DO UPDATE
				SET expires_at = excluded.expires_at;`,
		id, expiry.UTC().Unix(),
	)

	return errors.Wrap(err, "storing counter in database")
}

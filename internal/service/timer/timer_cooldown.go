package timer

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

// AddCooldown adds a new cooldown timer
func (s Service) AddCooldown(tt plugins.TimerType, limiter, ruleID string, expiry time.Time) error {
	return s.SetTimer(s.getCooldownTimerKey(tt, limiter, ruleID), expiry)
}

// InCooldown checks whether the cooldown has expired
func (s Service) InCooldown(tt plugins.TimerType, limiter, ruleID string) (bool, error) {
	return s.HasTimer(s.getCooldownTimerKey(tt, limiter, ruleID))
}

func (Service) getCooldownTimerKey(tt plugins.TimerType, limiter, ruleID string) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(fmt.Appendf(nil, "%d:%s:%s", tt, limiter, ruleID)))
}

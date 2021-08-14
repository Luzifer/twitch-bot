package main

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/Luzifer/twitch-bot/plugins"
)

var timerStore plugins.TimerStore = newTimer()

type timer struct{}

func newTimer() *timer {
	return &timer{}
}

// Cooldown timer

func (t *timer) AddCooldown(tt plugins.TimerType, limiter, ruleID string, expiry time.Time) {
	store.SetTimer(plugins.TimerTypeCooldown, t.getCooldownTimerKey(tt, limiter, ruleID), expiry)
}

func (t *timer) InCooldown(tt plugins.TimerType, limiter, ruleID string) bool {
	return store.HasTimer(t.getCooldownTimerKey(tt, limiter, ruleID))
}

func (timer) getCooldownTimerKey(tt plugins.TimerType, limiter, ruleID string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%d:%s:%s", tt, limiter, ruleID)
	return fmt.Sprintf("sha256:%x", h.Sum(nil))
}

// Permit timer

func (t *timer) AddPermit(channel, username string) {
	store.SetTimer(plugins.TimerTypePermit, t.getPermitTimerKey(channel, username), time.Now().Add(config.PermitTimeout))
}

func (t *timer) HasPermit(channel, username string) bool {
	return store.HasTimer(t.getPermitTimerKey(channel, username))
}

func (timer) getPermitTimerKey(channel, username string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%d:%s:%s", plugins.TimerTypePermit, channel, strings.ToLower(strings.TrimLeft(username, "@")))
	return fmt.Sprintf("sha256:%x", h.Sum(nil))
}

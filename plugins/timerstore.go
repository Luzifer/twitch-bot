package plugins

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

type TimerType uint8

const (
	TimerTypePermit TimerType = iota
	TimerTypeCooldown
)

type (
	TimerEntry struct {
		Kind TimerType `json:"kind"`
		Time time.Time `json:"time"`
	}

	TimerStore interface {
		AddCooldown(tt TimerType, limiter, ruleID string, expiry time.Time)
		InCooldown(tt TimerType, limiter, ruleID string) bool
		AddPermit(channel, username string)
		HasPermit(channel, username string) bool
	}

	testTimerStore struct {
		timers map[string]time.Time
	}
)

func newTestTimerStore() *testTimerStore { return &testTimerStore{timers: map[string]time.Time{}} }

// Cooldown timer

func (t *testTimerStore) AddCooldown(tt TimerType, limiter, ruleID string, expiry time.Time) {
	t.timers[t.getCooldownTimerKey(tt, limiter, ruleID)] = expiry
}

func (t *testTimerStore) InCooldown(tt TimerType, limiter, ruleID string) bool {
	return t.timers[t.getCooldownTimerKey(tt, limiter, ruleID)].After(time.Now())
}

func (testTimerStore) getCooldownTimerKey(tt TimerType, limiter, ruleID string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%d:%s:%s", tt, limiter, ruleID)
	return fmt.Sprintf("sha256:%x", h.Sum(nil))
}

// Permit timer

func (t *testTimerStore) AddPermit(channel, username string) {
	t.timers[t.getPermitTimerKey(channel, username)] = time.Now().Add(time.Minute)
}

func (t *testTimerStore) HasPermit(channel, username string) bool {
	return t.timers[t.getPermitTimerKey(channel, username)].After(time.Now())
}

func (testTimerStore) getPermitTimerKey(channel, username string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%d:%s:%s", TimerTypePermit, channel, strings.ToLower(strings.TrimLeft(username, "@")))
	return fmt.Sprintf("sha256:%x", h.Sum(nil))
}

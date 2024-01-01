package plugins

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

// TimerType defines an enum of available timer types
type TimerType uint8

// Definitions of supported TimerType values
const (
	TimerTypePermit TimerType = iota
	TimerTypeCooldown
)

type (
	// TimerEntry represents a time for the given type in the TimerStore
	TimerEntry struct {
		Kind TimerType `json:"kind"`
		Time time.Time `json:"time"`
	}

	// TimerStore defines what to expect when interacting with a store
	TimerStore interface {
		AddCooldown(tt TimerType, limiter, ruleID string, expiry time.Time) error
		InCooldown(tt TimerType, limiter, ruleID string) (bool, error)
		AddPermit(channel, username string) error
		HasPermit(channel, username string) (bool, error)
	}

	testTimerStore struct {
		timers map[string]time.Time
	}
)

var _ TimerStore = (*testTimerStore)(nil)

func newTestTimerStore() *testTimerStore { return &testTimerStore{timers: map[string]time.Time{}} }

// Cooldown timer

func (t *testTimerStore) AddCooldown(tt TimerType, limiter, ruleID string, expiry time.Time) error {
	t.timers[t.getCooldownTimerKey(tt, limiter, ruleID)] = expiry
	return nil
}

func (t *testTimerStore) InCooldown(tt TimerType, limiter, ruleID string) (bool, error) {
	return t.timers[t.getCooldownTimerKey(tt, limiter, ruleID)].After(time.Now()), nil
}

func (testTimerStore) getCooldownTimerKey(tt TimerType, limiter, ruleID string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%d:%s:%s", tt, limiter, ruleID)
	return fmt.Sprintf("sha256:%x", h.Sum(nil))
}

// Permit timer

func (t *testTimerStore) AddPermit(channel, username string) error {
	t.timers[t.getPermitTimerKey(channel, username)] = time.Now().Add(time.Minute)
	return nil
}

func (t *testTimerStore) HasPermit(channel, username string) (bool, error) {
	return t.timers[t.getPermitTimerKey(channel, username)].After(time.Now()), nil
}

func (testTimerStore) getPermitTimerKey(channel, username string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%d:%s:%s", TimerTypePermit, channel, strings.ToLower(strings.TrimLeft(username, "@")))
	return fmt.Sprintf("sha256:%x", h.Sum(nil))
}

package plugins

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

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
		// AddCooldown stores a cooldown timer of the given type until expiry.
		AddCooldown(tt TimerType, limiter, ruleID string, expiry time.Time) error

		// InCooldown reports whether the given limiter and rule are currently in cooldown.
		InCooldown(tt TimerType, limiter, ruleID string) (bool, error)

		// AddPermit stores a temporary permit for the given channel and username.
		AddPermit(channel, username string) error

		// HasPermit reports whether the given user currently has a permit in the channel.
		HasPermit(channel, username string) (bool, error)
	}

	// TimerType defines an enum of available timer types
	TimerType uint8

	testTimerStore struct {
		timers map[string]time.Time
	}
)

var _ TimerStore = (*testTimerStore)(nil)

func newTestTimerStore() *testTimerStore { return &testTimerStore{timers: make(map[string]time.Time)} }

func (t *testTimerStore) AddCooldown(tt TimerType, limiter, ruleID string, expiry time.Time) error {
	t.timers[t.getCooldownTimerKey(tt, limiter, ruleID)] = expiry
	return nil
}

func (t *testTimerStore) AddPermit(channel, username string) error {
	t.timers[t.getPermitTimerKey(channel, username)] = time.Now().Add(time.Minute)
	return nil
}

func (t *testTimerStore) HasPermit(channel, username string) (bool, error) {
	return t.timers[t.getPermitTimerKey(channel, username)].After(time.Now()), nil
}

func (t *testTimerStore) InCooldown(tt TimerType, limiter, ruleID string) (bool, error) {
	return t.timers[t.getCooldownTimerKey(tt, limiter, ruleID)].After(time.Now()), nil
}

func (testTimerStore) getCooldownTimerKey(tt TimerType, limiter, ruleID string) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(fmt.Appendf(nil, "%d:%s:%s", tt, limiter, ruleID)))
}

func (testTimerStore) getPermitTimerKey(channel, username string) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(fmt.Appendf(nil, "%d:%s:%s", TimerTypePermit, channel, strings.ToLower(strings.TrimLeft(username, "@")))))
}

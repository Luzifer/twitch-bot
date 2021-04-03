package main

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"
)

type timerType uint8

const (
	timerTypePermit timerType = iota
	timerTypeCooldown
)

var timerStore = newTimer()

type timerEntry struct {
	kind timerType
	time time.Time
}

type timer struct {
	timers map[string]timerEntry
	lock   *sync.RWMutex
}

func newTimer() *timer {
	return &timer{
		timers: map[string]timerEntry{},
		lock:   new(sync.RWMutex),
	}
}

// Cooldown timer

func (t *timer) AddCooldown(ruleID string) {
	t.add(timerTypeCooldown, t.getCooldownTimerKey(ruleID))
}

func (t *timer) InCooldown(ruleID string, cooldown time.Duration) bool {
	return t.has(t.getCooldownTimerKey(ruleID), cooldown)
}

func (t timer) getCooldownTimerKey(ruleID string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%d:%s", timerTypeCooldown, ruleID)
	return fmt.Sprintf("sha256:%x", h.Sum(nil))
}

// Permit timer

func (t *timer) AddPermit(channel, username string) {
	t.add(timerTypePermit, t.getPermitTimerKey(channel, username))
}

func (t *timer) HasPermit(channel, username string) bool {
	return t.has(t.getPermitTimerKey(channel, username), config.PermitTimeout)
}

func (t timer) getPermitTimerKey(channel, username string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%d:%s:%s", timerTypePermit, channel, strings.ToLower(strings.TrimLeft(username, "@")))
	return fmt.Sprintf("sha256:%x", h.Sum(nil))
}

// Generic

func (t *timer) add(kind timerType, id string) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.timers[id] = timerEntry{kind: kind, time: time.Now()}
}

func (t *timer) has(id string, validity time.Duration) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return time.Since(t.timers[id].time) < validity
}

package main

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

type timerType uint8

const (
	timerTypePermit timerType = iota
	timerTypeCooldown
)

var timerStore = newTimer()

type timerEntry struct {
	Kind timerType `json:"kind"`
	Time time.Time `json:"time"`
}

type timer struct{}

func newTimer() *timer {
	return &timer{}
}

// Cooldown timer

func (t *timer) AddCooldown(tt timerType, limiter, ruleID string, expiry time.Time) {
	store.SetTimer(timerTypeCooldown, t.getCooldownTimerKey(tt, limiter, ruleID), expiry)
}

func (t *timer) InCooldown(tt timerType, limiter, ruleID string) bool {
	return store.HasTimer(t.getCooldownTimerKey(tt, limiter, ruleID))
}

func (timer) getCooldownTimerKey(tt timerType, limiter, ruleID string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%d:%s:%s", tt, limiter, ruleID)
	return fmt.Sprintf("sha256:%x", h.Sum(nil))
}

// Permit timer

func (t *timer) AddPermit(channel, username string) {
	store.SetTimer(timerTypePermit, t.getPermitTimerKey(channel, username), time.Now().Add(config.PermitTimeout))
}

func (t *timer) HasPermit(channel, username string) bool {
	return store.HasTimer(t.getPermitTimerKey(channel, username))
}

func (timer) getPermitTimerKey(channel, username string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%d:%s:%s", timerTypePermit, channel, strings.ToLower(strings.TrimLeft(username, "@")))
	return fmt.Sprintf("sha256:%x", h.Sum(nil))
}

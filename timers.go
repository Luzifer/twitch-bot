package main

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"
)

var timerStore = newTimer()

type timer struct {
	timers map[string]time.Time
	lock   *sync.RWMutex
}

func newTimer() *timer {
	return &timer{
		timers: map[string]time.Time{},
		lock:   new(sync.RWMutex),
	}
}

func (t *timer) Add(id string) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.timers[id] = time.Now()
}

func (t *timer) AddPermit(channel, username string) {
	t.Add(t.getPermitTimerKey(channel, username))
}

func (t *timer) Has(id string, validity time.Duration) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return time.Since(t.timers[id]) < validity
}

func (t *timer) HasPermit(channel, username string) bool {
	return t.Has(t.getPermitTimerKey(channel, username), config.PermitTimeout)
}

func (t timer) NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimLeft(username, "@"))
}

func (t timer) getPermitTimerKey(channel, username string) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s:%s", channel, t.NormalizeUsername(username))
	return fmt.Sprintf("sha256:%x", h.Sum(nil))
}

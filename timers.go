package main

import (
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

func (t *timer) Has(id string, validity time.Duration) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return time.Since(t.timers[id]) < validity
}

func (t *timer) HasPermit(username string) bool {
	return t.Has(t.NormalizeUsername(username), config.PermitTimeout)
}

func (t timer) NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimLeft(username, "@"))
}

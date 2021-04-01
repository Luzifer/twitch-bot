package main

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"
)

type (
	twitchAPICache struct {
		data map[string]twitchAPICacheEntry
		lock sync.RWMutex
	}
	twitchAPICacheEntry struct {
		Data       interface{}
		ValidUntil time.Time
	}
)

func newTwitchAPICache() *twitchAPICache {
	return &twitchAPICache{
		data: make(map[string]twitchAPICacheEntry),
	}
}

func (t *twitchAPICache) Get(key []string) interface{} {
	t.lock.RLock()
	defer t.lock.RUnlock()

	e := t.data[t.deriveKey(key)]
	if e.ValidUntil.Before(time.Now()) {
		return nil
	}

	return e.Data
}

func (t *twitchAPICache) Set(key []string, valid time.Duration, data interface{}) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.data[t.deriveKey(key)] = twitchAPICacheEntry{
		Data:       data,
		ValidUntil: time.Now().Add(valid),
	}
}

func (*twitchAPICache) deriveKey(key []string) string {
	sha := sha256.New()

	fmt.Fprintf(sha, "%s", strings.Join(key, ":"))

	return fmt.Sprintf("sha256:%x", sha.Sum(nil))
}

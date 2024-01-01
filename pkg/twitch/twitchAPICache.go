package twitch

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"
)

type (
	// APICache is used to cache API responses in order not to ask the
	// API over and over again for information seldom changing
	APICache struct {
		data map[string]twitchAPICacheEntry
		lock sync.RWMutex
	}

	twitchAPICacheEntry struct {
		Data       interface{}
		ValidUntil time.Time
	}
)

func newTwitchAPICache() *APICache {
	return &APICache{
		data: make(map[string]twitchAPICacheEntry),
	}
}

// Get returns the stored data or nil for the given cache-key
func (t *APICache) Get(key []string) interface{} {
	t.lock.RLock()
	defer t.lock.RUnlock()

	e := t.data[t.deriveKey(key)]
	if e.ValidUntil.Before(time.Now()) {
		return nil
	}

	return e.Data
}

// Set sets the stored data for the given cache-key
func (t *APICache) Set(key []string, valid time.Duration, data interface{}) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.data[t.deriveKey(key)] = twitchAPICacheEntry{
		Data:       data,
		ValidUntil: time.Now().Add(valid),
	}
}

func (*APICache) deriveKey(key []string) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(strings.Join(key, ":"))))
}

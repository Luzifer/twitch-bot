package twitch

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"
)

type (
	TwitchAPICache struct {
		data map[string]twitchAPICacheEntry
		lock sync.RWMutex
	}

	twitchAPICacheEntry struct {
		Data       interface{}
		ValidUntil time.Time
	}
)

func newTwitchAPICache() *TwitchAPICache {
	return &TwitchAPICache{
		data: make(map[string]twitchAPICacheEntry),
	}
}

func (t *TwitchAPICache) Get(key []string) interface{} {
	t.lock.RLock()
	defer t.lock.RUnlock()

	e := t.data[t.deriveKey(key)]
	if e.ValidUntil.Before(time.Now()) {
		return nil
	}

	return e.Data
}

func (t *TwitchAPICache) Set(key []string, valid time.Duration, data interface{}) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.data[t.deriveKey(key)] = twitchAPICacheEntry{
		Data:       data,
		ValidUntil: time.Now().Add(valid),
	}
}

func (*TwitchAPICache) deriveKey(key []string) string {
	sha := sha256.New()

	fmt.Fprintf(sha, "%s", strings.Join(key, ":"))

	return fmt.Sprintf("sha256:%x", sha.Sum(nil))
}

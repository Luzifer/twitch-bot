package main

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

type (
	twitchAPICache      map[string]twitchAPICacheEntry
	twitchAPICacheEntry struct {
		Data       interface{}
		ValidUntil time.Time
	}
)

func (t twitchAPICache) Get(key []string) interface{} {
	e := t[t.deriveKey(key)]
	if e.ValidUntil.Before(time.Now()) {
		return nil
	}

	return e.Data
}

func (t twitchAPICache) Set(key []string, valid time.Duration, data interface{}) {
	t[t.deriveKey(key)] = twitchAPICacheEntry{
		Data:       data,
		ValidUntil: time.Now().Add(valid),
	}
}

func (twitchAPICache) deriveKey(key []string) string {
	sha := sha256.New()

	fmt.Fprintf(sha, "%s", strings.Join(key, ":"))

	return fmt.Sprintf("sha256:%x", sha.Sum(nil))
}

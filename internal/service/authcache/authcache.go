// Package authcache implements a cache for token auth to hold auth-
// results with cpu/mem inexpensive methods instead of always using
// secure but expensive methods to validate the token
package authcache

import (
	"crypto/sha256"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const negativeCacheTime = 5 * time.Minute

// AuthBackendAny signals any backend can answer the validity of the
// token and no specific backend is selected
const AuthBackendAny = "*"

type (
	// Service manages the cached auth results
	Service struct {
		backends     map[string]AuthFunc
		backendOrder []string
		cache        map[string]*CacheEntry
		lock         sync.RWMutex
	}

	// CacheEntry represents an entry in the cache Service
	CacheEntry struct {
		AuthResult error // Allows for negative caching
		ExpiresAt  time.Time
		Modules    []string
	}

	// CreateOpt supplies options to the creation of the cache
	CreateOpt func(*Service)

	// AuthFunc is a backend-function to resolve a token to a list of
	// modules the token is authorized for, an expiry-time and an error.
	// The error MUST be ErrUnauthorized in case the user is not found,
	// if the error is another, the backend resolve will be cancelled
	// and no further backends are queried.
	AuthFunc func(token string) (modules []string, expiresAt time.Time, err error)
)

// ErrUnauthorized denotes the token could not be found in any backend
// auth method and therefore is not an user
var ErrUnauthorized = errors.New("unauthorized")

// New creates a new Service with the given backend methods to
// authenticate users
func New(fns ...CreateOpt) *Service {
	s := &Service{
		backends: make(map[string]AuthFunc),
		cache:    make(map[string]*CacheEntry),
	}

	for _, fn := range fns {
		fn(s)
	}

	go s.runCleanup()

	return s
}

// WithAuthBackend adds a new auth backend which is tried in order of
// begin added
func WithAuthBackend(name string, fn AuthFunc) CreateOpt {
	return func(s *Service) {
		s.backendOrder = append(s.backendOrder, name)
		s.backends[name] = fn
	}
}

// ValidateTokenFor checks backends whether the given token has access
// to the given modules and caches the result
func (s *Service) ValidateTokenFor(token string, modules ...string) error {
	return s.ValidateTokenForWithTokenType(token, AuthBackendAny, modules...)
}

// ValidateTokenForWithTokenType checks the configured backends whether the
// given token has access to the given modules and caches the result for the
// selected auth backend name.
func (s *Service) ValidateTokenForWithTokenType(token string, authBackendName string, modules ...string) error {
	s.lock.RLock()
	cached := s.cache[s.cacheKey(token, authBackendName)]
	s.lock.RUnlock()

	if cached != nil && cached.ExpiresAt.After(time.Now()) {
		// We do have a recent cache entry for that token: continue to use
		return cached.validateFor(modules)
	}

	// No recent cache entry: We need to ask the expensive backends
	var (
		ce              CacheEntry
		backendSelected bool
	)
backendLoop:
	for _, backendName := range s.backendOrder {
		if authBackendName != AuthBackendAny && backendName != authBackendName {
			continue
		}

		backendSelected = true
		ce.Modules, ce.ExpiresAt, ce.AuthResult = s.backends[backendName](token)
		switch {
		case ce.AuthResult == nil:
			// Valid result & auth, the user was found
			break backendLoop

		case errors.Is(ce.AuthResult, ErrUnauthorized):
			// Valid result, user was not found
			continue backendLoop

		default:
			// Something went wrong, bail out and do not cache
			return errors.Wrap(ce.AuthResult, "querying authorization in backend")
		}
	}

	if !backendSelected {
		ce.AuthResult = ErrUnauthorized
	}

	// We got a final result: That might be ErrUnauthorized or a valid
	// user. Both should be cached. The error for a static time, the
	// valid result for the time given by the backend.
	if errors.Is(ce.AuthResult, ErrUnauthorized) {
		ce.ExpiresAt = time.Now().Add(negativeCacheTime)
	}

	s.lock.Lock()
	s.cache[s.cacheKey(token, authBackendName)] = &ce
	s.lock.Unlock()

	// Finally return the result for the requested modules
	return ce.validateFor(modules)
}

func (*Service) cacheKey(token string, backendName string) string {
	return fmt.Sprintf("sha256:%s:%x", backendName, sha256.Sum256([]byte(token)))
}

func (s *Service) cleanup() {
	s.lock.Lock()
	defer s.lock.Unlock()

	var (
		now    = time.Now()
		remove []string
	)

	for key := range s.cache {
		if s.cache[key].ExpiresAt.Before(now) {
			remove = append(remove, key)
		}
	}

	for _, key := range remove {
		delete(s.cache, key)
	}
}

func (s *Service) runCleanup() {
	for range time.NewTicker(time.Minute).C {
		s.cleanup()
	}
}

func (c CacheEntry) validateFor(modules []string) error {
	if c.AuthResult != nil {
		return c.AuthResult
	}

	for _, reqMod := range modules {
		if !slices.Contains(c.Modules, reqMod) && !slices.Contains(c.Modules, "*") {
			return errors.New("missing module in auth")
		}
	}

	return nil
}

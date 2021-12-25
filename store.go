package main

import (
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/Luzifer/twitch-bot/plugins"
)

const eventSubSecretLength = 32

type storageFile struct {
	Counters  map[string]int64              `json:"counters"`
	Timers    map[string]plugins.TimerEntry `json:"timers"`
	Variables map[string]string             `json:"variables"`

	ModuleStorage map[string]json.RawMessage `json:"module_storage"`

	GrantedScopes map[string][]string `json:"granted_scopes"`

	EventSubSecret string `json:"event_sub_secret,omitempty"`

	inMem bool
	lock  *sync.RWMutex
}

func newStorageFile(inMemStore bool) *storageFile {
	return &storageFile{
		Counters:  map[string]int64{},
		Timers:    map[string]plugins.TimerEntry{},
		Variables: map[string]string{},

		ModuleStorage: map[string]json.RawMessage{},

		GrantedScopes: map[string][]string{},

		inMem: inMemStore,
		lock:  new(sync.RWMutex),
	}
}

func (s *storageFile) DeleteGrantedScopes(user string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.GrantedScopes, user)

	return errors.Wrap(s.Save(), "saving store")
}

func (s *storageFile) DeleteModuleStore(moduleUUID string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.ModuleStorage, moduleUUID)

	return errors.Wrap(s.Save(), "saving store")
}

func (s *storageFile) GetCounterValue(counter string) int64 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.Counters[counter]
}

func (s *storageFile) GetOrGenerateEventSubSecret() (string, string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.EventSubSecret != "" {
		return s.EventSubSecret, s.EventSubSecret[:5], nil
	}

	key := make([]byte, eventSubSecretLength)
	n, err := rand.Read(key)
	if err != nil {
		return "", "", errors.Wrap(err, "generating random secret")
	}
	if n != eventSubSecretLength {
		return "", "", errors.Errorf("read only %d of %d byte", n, eventSubSecretLength)
	}

	s.EventSubSecret = hex.EncodeToString(key)

	return s.EventSubSecret, s.EventSubSecret[:5], errors.Wrap(s.Save(), "saving store")
}

func (s *storageFile) GetModuleStore(moduleUUID string, storedObject plugins.StorageUnmarshaller) error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return errors.Wrap(
		storedObject.UnmarshalStoredObject(s.ModuleStorage[moduleUUID]),
		"unmarshalling stored object",
	)
}

func (s *storageFile) GetVariable(key string) string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.Variables[key]
}

func (s *storageFile) HasTimer(id string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.Timers[id].Time.After(time.Now())
}

func (s *storageFile) Load() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.inMem {
		// In-Memory store is active, do not load from disk
		// for testing purposes only!
		return nil
	}

	f, err := os.Open(cfg.StorageFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Store init state
			return nil
		}
		return errors.Wrap(err, "open storage file")
	}
	defer f.Close()

	zf, err := gzip.NewReader(f)
	if err != nil {
		return errors.Wrap(err, "create gzip reader")
	}
	defer zf.Close()

	return errors.Wrap(
		json.NewDecoder(zf).Decode(s),
		"decode storage object",
	)
}

func (s *storageFile) Save() error {
	// NOTE(kahlers): DO NOT LOCK THIS, all calling functions are
	// modifying functions and must have locks in place

	if s.inMem {
		// In-Memory store is active, do not store to disk
		// for testing purposes only!
		return nil
	}

	// Cleanup timers
	var timerIDs []string
	for id := range s.Timers {
		timerIDs = append(timerIDs, id)
	}

	for _, i := range timerIDs {
		if s.Timers[i].Time.Before(time.Now()) {
			delete(s.Timers, i)
		}
	}

	// Write store to disk
	f, err := os.Create(cfg.StorageFile)
	if err != nil {
		return errors.Wrap(err, "create storage file")
	}
	defer f.Close()

	zf := gzip.NewWriter(f)
	defer zf.Close()

	return errors.Wrap(
		json.NewEncoder(zf).Encode(s),
		"encode storage object",
	)
}

func (s *storageFile) SetGrantedScopes(user string, scopes []string, merge bool) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if merge {
		for _, sc := range s.GrantedScopes[user] {
			if !str.StringInSlice(sc, scopes) {
				scopes = append(scopes, sc)
			}
		}
	}

	s.GrantedScopes[user] = scopes

	return errors.Wrap(s.Save(), "saving store")
}

func (s *storageFile) SetModuleStore(moduleUUID string, storedObject plugins.StorageMarshaller) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	data, err := storedObject.MarshalStoredObject()
	if err != nil {
		return errors.Wrap(err, "marshalling stored object")
	}

	s.ModuleStorage[moduleUUID] = data

	return errors.Wrap(s.Save(), "saving store")
}

func (s *storageFile) SetTimer(kind plugins.TimerType, id string, expiry time.Time) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.Timers[id] = plugins.TimerEntry{Kind: kind, Time: expiry}

	return errors.Wrap(s.Save(), "saving store")
}

func (s *storageFile) SetVariable(key, value string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.Variables[key] = value

	return errors.Wrap(s.Save(), "saving store")
}

func (s *storageFile) RemoveVariable(key string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.Variables, key)

	return errors.Wrap(s.Save(), "saving store")
}

func (s *storageFile) UpdateCounter(counter string, value int64, absolute bool) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !absolute {
		value = s.Counters[counter] + value
	}

	s.Counters[counter] = value

	return errors.Wrap(s.Save(), "saving store")
}

func (s *storageFile) UserHasGrantedAnyScope(user string, scopes ...string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	grantedScopes, ok := s.GrantedScopes[user]
	if !ok {
		return false
	}

	for _, scope := range scopes {
		if str.StringInSlice(scope, grantedScopes) {
			return true
		}
	}

	return false
}

func (s *storageFile) UserHasGrantedScopes(user string, scopes ...string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	grantedScopes, ok := s.GrantedScopes[user]
	if !ok {
		return false
	}

	for _, scope := range scopes {
		if !str.StringInSlice(scope, grantedScopes) {
			return false
		}
	}

	return true
}

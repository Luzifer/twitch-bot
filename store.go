package main

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type storageFile struct {
	Counters  map[string]int64      `json:"counters"`
	Timers    map[string]timerEntry `json:"timers"`
	Variables map[string]string     `json:"variables"`

	inMem bool
	lock  *sync.RWMutex
}

func newStorageFile(inMemStore bool) *storageFile {
	return &storageFile{
		Counters:  map[string]int64{},
		Timers:    map[string]timerEntry{},
		Variables: map[string]string{},

		inMem: inMemStore,
		lock:  new(sync.RWMutex),
	}
}

func (s *storageFile) GetCounterValue(counter string) int64 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.Counters[counter]
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

func (s *storageFile) SetTimer(kind timerType, id string, expiry time.Time) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.Timers[id] = timerEntry{Kind: kind, Time: expiry}

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

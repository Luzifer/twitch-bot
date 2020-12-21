package main

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"sync"

	"github.com/pkg/errors"
)

type storageFile struct {
	Counters map[string]int64 `json:"counters"`

	lock *sync.RWMutex
}

func newStorageFile() *storageFile {
	return &storageFile{
		Counters: map[string]int64{},

		lock: new(sync.RWMutex),
	}
}

func (s *storageFile) GetCounterValue(counter string) int64 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.Counters[counter]
}

func (s *storageFile) Load() error {
	s.lock.Lock()
	defer s.lock.Unlock()

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

func (s *storageFile) UpdateCounter(counter string, value int64, absolute bool) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !absolute {
		value = s.Counters[counter] + value
	}

	s.Counters[counter] = value

	return errors.Wrap(s.Save(), "saving store")
}

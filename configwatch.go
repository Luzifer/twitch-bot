package main

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

const configChangeCheckInterval = time.Second

type configChangeEvent uint8

const (
	configChangeEventUnkown configChangeEvent = iota
	configChangeEventNotExist
	configChangeEventModified
)

func watchConfigChanges(filename string, evt chan configChangeEvent) {
	var (
		available   bool
		initialized bool
		size        int64
		modTime     time.Time
	)

	for range time.NewTicker(configChangeCheckInterval).C {
		info, err := os.Stat(filename)
		switch {
		case err == nil:
			// Fine

		case os.IsNotExist(err):
			if available {
				evt <- configChangeEventNotExist
			}
			available = false
			continue

		default:
			log.WithError(err).Error("Failed to get config stat")
			continue
		}

		if initialized && (info.Size() != size || !info.ModTime().Equal(modTime)) {
			evt <- configChangeEventModified
		}

		available = true
		initialized = true
		size = info.Size()
		modTime = info.ModTime()
	}
}

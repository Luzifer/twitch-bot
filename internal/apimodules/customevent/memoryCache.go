package customevent

import (
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v3/pkg/database"
)

const memoryCacheRefreshInterval = 5 * time.Minute

type (
	memoryCache struct {
		events     []storedCustomEvent
		validUntil time.Time

		dbc  database.Connector
		lock sync.Mutex
	}
)

func (m *memoryCache) PopEventsToExecute() ([]storedCustomEvent, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.validUntil.Before(time.Now()) {
		if err := m.refresh(); err != nil {
			return nil, errors.Wrap(err, "refreshing stale cache")
		}
	}

	var (
		execEvents, storeEvents []storedCustomEvent
		now                     = time.Now()
	)
	for i := range m.events {
		evt := m.events[i]
		if evt.ScheduledAt.After(now) {
			storeEvents = append(storeEvents, evt)
			continue
		}

		execEvents = append(execEvents, evt)
	}

	m.events = storeEvents
	return execEvents, nil
}

func (m *memoryCache) Refresh() (err error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.refresh()
}

func (m *memoryCache) refresh() (err error) {
	if m.events, err = getFutureEvents(m.dbc); err != nil {
		return errors.Wrap(err, "fetching events from database")
	}

	m.validUntil = time.Now().Add(memoryCacheRefreshInterval)
	logrus.WithField("event_count", len(m.events)).Trace("loaded stored events from database")
	return nil
}

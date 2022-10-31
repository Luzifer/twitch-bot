package customevent

import (
	"strings"

	"github.com/sirupsen/logrus"
)

func scheduleCleanup() {
	if err := cleanupStoredEvents(db); err != nil {
		logrus.WithError(err).Error("executing custom event database cleanup")
	}
}

func scheduleSend() {
	evts, err := mc.PopEventsToExecute()
	if err != nil {
		logrus.WithError(err).Error("collecting scheduled custom events for sending")
		return
	}

	for i := range evts {
		go func(evt storedCustomEvent) {
			evtData, err := parseEvent(evt.Channel, strings.NewReader(evt.Fields))
			if err != nil {
				logrus.WithError(err).Error("parsing fields in stored event")
				return
			}

			if err = eventCreatorFunc("custom", evtData); err != nil {
				logrus.WithError(err).Error("triggering stored event")
				return
			}
		}(evts[i])
	}
}

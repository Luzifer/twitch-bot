package customevent

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid/v3"
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v3/pkg/database"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const cleanupTimeout = 15 * time.Minute

type (
	storedCustomEvent struct {
		ID          string `gorm:"primaryKey"`
		Channel     string
		Fields      string
		ScheduledAt time.Time
	}
)

func cleanupStoredEvents(db database.Connector) error {
	return errors.Wrap(
		db.DB().
			Where("scheduled_at < ?", time.Now().Add(cleanupTimeout*-1).UTC()).
			Delete(&storedCustomEvent{}).
			Error,
		"deleting past events",
	)
}

func getFutureEvents(db database.Connector) (out []storedCustomEvent, err error) {
	return out, errors.Wrap(
		db.DB().
			Where("scheduled_at >= ?", time.Now().UTC()).
			Find(&out).
			Error,
		"getting events from database",
	)
}

func storeEvent(db database.Connector, scheduleAt time.Time, channel string, fields *plugins.FieldCollection) error {
	fieldBuf := new(bytes.Buffer)
	if err := json.NewEncoder(fieldBuf).Encode(fields); err != nil {
		return errors.Wrap(err, "marshalling fields")
	}

	return errors.Wrap(
		db.DB().Create(storedCustomEvent{
			ID:          uuid.Must(uuid.NewV4()).String(),
			Channel:     channel,
			Fields:      fieldBuf.String(),
			ScheduledAt: scheduleAt,
		}).Error,
		"storing event",
	)
}

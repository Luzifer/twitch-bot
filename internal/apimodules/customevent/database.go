package customevent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"github.com/gofrs/uuid/v3"
	"gorm.io/gorm"

	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/database"
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
	if err := helpers.RetryTransaction(db.DB(), func(tx *gorm.DB) error {
		return tx.Where("scheduled_at < ?", time.Now().Truncate(time.Second).Add(cleanupTimeout*-1).UTC()).
			Delete(&storedCustomEvent{}).
			Error
	}); err != nil {
		return fmt.Errorf("deleting past events: %w", err)
	}

	return nil
}

func getFutureEvents(db database.Connector) (out []storedCustomEvent, err error) {
	if err := helpers.Retry(func() error {
		return db.DB().
			Where("scheduled_at >= ?", time.Now().Truncate(time.Second).UTC()).
			Find(&out).
			Error
	}); err != nil {
		return nil, fmt.Errorf("getting events from database: %w", err)
	}

	return out, nil
}

func storeEvent(db database.Connector, scheduleAt time.Time, channel string, fields *fieldcollection.FieldCollection) error {
	fieldBuf := new(bytes.Buffer)
	if err := json.NewEncoder(fieldBuf).Encode(fields); err != nil {
		return fmt.Errorf("marshalling fields: %w", err)
	}

	if err := helpers.RetryTransaction(db.DB(), func(tx *gorm.DB) error {
		return tx.Create(storedCustomEvent{
			ID:          uuid.Must(uuid.NewV4()).String(),
			Channel:     channel,
			Fields:      fieldBuf.String(),
			ScheduledAt: scheduleAt.Truncate(time.Second),
		}).Error
	}); err != nil {
		return fmt.Errorf("storing event: %w", err)
	}

	return nil
}

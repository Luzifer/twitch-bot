package overlays

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/Luzifer/go_helpers/v2/backoff"
	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/database"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

type (
	overlaysEvent struct {
		ID        uint64    `gorm:"primaryKey"`
		Channel   string    `gorm:"not null;index:overlays_events_sort_idx"`
		CreatedAt time.Time `gorm:"index:overlays_events_sort_idx"`
		EventType string
		Fields    string
	}
)

func addChannelEvent(db database.Connector, channel string, evt socketMessage) (evtID uint64, err error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(evt.Fields); err != nil {
		return 0, errors.Wrap(err, "encoding fields")
	}

	storEvt := &overlaysEvent{
		Channel:   channel,
		CreatedAt: evt.Time.UTC(),
		EventType: evt.Type,
		Fields:    strings.TrimSpace(buf.String()),
	}

	if err = helpers.RetryTransaction(db.DB(), func(tx *gorm.DB) error {
		return tx.Create(storEvt).Error
	}); err != nil {
		return 0, errors.Wrap(err, "storing event to database")
	}

	return storEvt.ID, nil
}

func getChannelEvents(db database.Connector, channel string) ([]socketMessage, error) {
	var evts []overlaysEvent

	if err := helpers.Retry(func() error {
		return db.DB().Where("channel = ?", channel).Order("created_at").Find(&evts).Error
	}); err != nil {
		return nil, errors.Wrap(err, "querying channel events")
	}

	var out []socketMessage
	for _, e := range evts {
		sm, err := e.ToSocketMessage()
		if err != nil {
			return nil, errors.Wrap(err, "transforming event")
		}

		out = append(out, sm)
	}

	return out, nil
}

func getEventByID(db database.Connector, eventID uint64) (socketMessage, error) {
	var evt overlaysEvent

	if err := helpers.Retry(func() (err error) {
		err = db.DB().Where("id = ?", eventID).First(&evt).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return backoff.NewErrCannotRetry(err)
		}
		return err
	}); err != nil {
		return socketMessage{}, errors.Wrap(err, "fetching event")
	}

	return evt.ToSocketMessage()
}

func (o overlaysEvent) ToSocketMessage() (socketMessage, error) {
	fields := new(plugins.FieldCollection)
	if err := json.NewDecoder(strings.NewReader(o.Fields)).Decode(fields); err != nil {
		return socketMessage{}, errors.Wrap(err, "decoding fields")
	}

	return socketMessage{
		EventID: o.ID,
		IsLive:  false,
		Time:    o.CreatedAt,
		Type:    o.EventType,
		Fields:  fields,
	}, nil
}

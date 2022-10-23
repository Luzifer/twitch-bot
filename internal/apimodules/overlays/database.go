package overlays

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v2/pkg/database"
	"github.com/Luzifer/twitch-bot/v2/plugins"
)

type (
	overlaysEvent struct {
		Channel   string    `gorm:"not null;index:overlays_events_sort_idx"`
		CreatedAt time.Time `gorm:"index:overlays_events_sort_idx"`
		EventType string
		Fields    string
	}
)

func AddChannelEvent(db database.Connector, channel string, evt SocketMessage) error {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(evt.Fields); err != nil {
		return errors.Wrap(err, "encoding fields")
	}

	return errors.Wrap(
		db.DB().Create(overlaysEvent{
			Channel:   channel,
			CreatedAt: evt.Time.UTC(),
			EventType: evt.Type,
			Fields:    strings.TrimSpace(buf.String()),
		}).Error,
		"storing event to database",
	)
}

func GetChannelEvents(db database.Connector, channel string) ([]SocketMessage, error) {
	var evts []overlaysEvent

	if err := db.DB().Where("channel = ?", channel).Order("created_at").Find(&evts).Error; err != nil {
		return nil, errors.Wrap(err, "querying channel events")
	}

	var out []SocketMessage
	for _, e := range evts {
		fields := new(plugins.FieldCollection)
		if err := json.NewDecoder(strings.NewReader(e.Fields)).Decode(fields); err != nil {
			return nil, errors.Wrap(err, "decoding fields")
		}

		out = append(out, SocketMessage{
			IsLive: false,
			Time:   e.CreatedAt,
			Type:   e.EventType,
			Fields: fields,
		})
	}

	return out, nil
}

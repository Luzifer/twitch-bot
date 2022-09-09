package v2migrator

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/internal/database"
	"github.com/Luzifer/twitch-bot/plugins"
)

type (
	storageModOverlays struct {
		ChannelEvents map[string][]struct {
			IsLive bool                     `json:"is_live"`
			Time   time.Time                `json:"time"`
			Type   string                   `json:"type"`
			Fields *plugins.FieldCollection `json:"fields"`
		} `json:"channel_events"`
	}
)

func (s storageModOverlays) migrate(db database.Connector) (err error) {
	for channel, evts := range s.ChannelEvents {
		for _, evt := range evts {
			buf := new(bytes.Buffer)
			if err = json.NewEncoder(buf).Encode(evt.Fields); err != nil {
				return errors.Wrap(err, "encoding fields")
			}

			if _, err = db.DB().Exec(
				`INSERT INTO overlays_events
					(channel, created_at, event_type, fields)
					VALUES ($1, $2, $3, $4);`,
				channel, evt.Time.UnixNano(), evt.Type, strings.TrimSpace(buf.String()),
			); err != nil {
				return errors.Wrap(err, "storing event to database")
			}
		}
	}

	return nil
}

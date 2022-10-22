package v2migrator

import (
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/internal/apimodules/overlays"
	"github.com/Luzifer/twitch-bot/pkg/database"
)

type (
	storageModOverlays struct {
		ChannelEvents map[string][]overlays.SocketMessage `json:"channel_events"`
	}
)

func (s storageModOverlays) migrate(db database.Connector) (err error) {
	for channel, evts := range s.ChannelEvents {
		for _, evt := range evts {
			if err := overlays.AddChannelEvent(db, channel, evt); err != nil {
				return errors.Wrap(err, "storing event to database")
			}
		}
	}

	return nil
}

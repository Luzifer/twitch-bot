package v2migrator

import (
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v2/internal/actors/quotedb"
	"github.com/Luzifer/twitch-bot/v2/pkg/database"
)

type (
	storageModQuoteDB struct {
		ChannelQuotes map[string][]string `json:"channel_quotes"`
	}
)

func (s storageModQuoteDB) migrate(db database.Connector) (err error) {
	for channel, quotes := range s.ChannelQuotes {
		if err := quotedb.SetQuotes(db, channel, quotes); err != nil {
			return errors.Wrap(err, "setting quotes for channel")
		}
	}

	return nil
}

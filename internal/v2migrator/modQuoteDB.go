package v2migrator

import (
	"time"

	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/pkg/database"
)

type (
	storageModQuoteDB struct {
		ChannelQuotes map[string][]string `json:"channel_quotes"`
	}
)

func (s storageModQuoteDB) migrate(db database.Connector) (err error) {
	for channel, quotes := range s.ChannelQuotes {
		t := time.Now()
		for _, quote := range quotes {
			if _, err = db.DB().Exec(
				`INSERT INTO quotedb
					(channel, created_at, quote)
					VALUES ($1, $2, $3);`,
				channel, t.UnixNano(), quote,
			); err != nil {
				return errors.Wrap(err, "adding quote for channel")
			}

			t = t.Add(time.Nanosecond) // Increase by one ns to adhere to unique index
		}
	}

	return nil
}

package v2migrator

import (
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/internal/service/access"
	"github.com/Luzifer/twitch-bot/pkg/database"
)

func (s storageFile) migrateCoreKV(db database.Connector) (err error) {
	as := access.New(db)

	if err = as.SetBotTwitchCredentials(s.BotAccessToken, s.BotRefreshToken); err != nil {
		return errors.Wrap(err, "setting bot credentials")
	}

	if err = db.StoreCoreMeta("event_sub_secret", s.EventSubSecret); err != nil {
		return errors.Wrap(err, "storing bot eventsub token")
	}

	return nil
}

func (s storageFile) migrateCounters(db database.Connector) (err error) {
	for counter, value := range s.Counters {
		if _, err = db.DB().Exec(
			`INSERT INTO counters
				(name, value)
				VALUES ($1, $2)
				ON CONFLICT DO UPDATE
					SET value = excluded.value;`,
			counter, value,
		); err != nil {
			return errors.Wrap(err, "storing counter value")
		}
	}

	return nil
}

func (s storageFile) migratePermissions(db database.Connector) (err error) {
	as := access.New(db)

	for channel, perms := range s.ExtendedPermissions {
		if err = as.SetExtendedTwitchCredentials(
			channel,
			perms.AccessToken,
			perms.RefreshToken,
			perms.Scopes,
		); err != nil {
			return errors.Wrapf(err, "storing channel %q credentials", channel)
		}
	}

	return nil
}

func (s storageFile) migrateTimers(db database.Connector) (err error) {
	for id, expiry := range s.Timers {
		if _, err := db.DB().Exec(
			`INSERT INTO timers
				(id, expires_at)
				VALUES ($1, $2)
				ON CONFLICT DO UPDATE
					SET expires_at = excluded.expires_at;`,
			id, expiry.Time.Unix(),
		); err != nil {
			return errors.Wrap(err, "storing counter in database")
		}
	}

	return nil
}

func (s storageFile) migrateVariables(db database.Connector) (err error) {
	for key, value := range s.Variables {
		if _, err = db.DB().Exec(
			`INSERT INTO variables
				(name, value)
				VALUES ($1, $2)
				ON CONFLICT DO UPDATE
					SET value = excluded.value;`,
			key, value,
		); err != nil {
			return errors.Wrap(err, "updating value in database")
		}
	}

	return nil
}

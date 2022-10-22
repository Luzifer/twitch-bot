package v2migrator

import (
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/internal/actors/counter"
	"github.com/Luzifer/twitch-bot/internal/actors/variables"
	"github.com/Luzifer/twitch-bot/internal/service/access"
	"github.com/Luzifer/twitch-bot/internal/service/timer"
	"github.com/Luzifer/twitch-bot/pkg/database"
)

func (s storageFile) migrateCoreKV(db database.Connector) (err error) {
	as, err := access.New(db)
	if err != nil {
		return errors.Wrap(err, "creating access service")
	}

	if err = as.SetBotTwitchCredentials(s.BotAccessToken, s.BotRefreshToken); err != nil {
		return errors.Wrap(err, "setting bot credentials")
	}

	if err = db.StoreEncryptedCoreMeta("event_sub_secret", s.EventSubSecret); err != nil {
		return errors.Wrap(err, "storing bot eventsub token")
	}

	return nil
}

func (s storageFile) migrateCounters(db database.Connector) (err error) {
	for counterName, value := range s.Counters {
		if err = counter.UpdateCounter(db, counterName, value, true); err != nil {
			return errors.Wrap(err, "storing counter value")
		}
	}

	return nil
}

func (s storageFile) migratePermissions(db database.Connector) (err error) {
	as, err := access.New(db)
	if err != nil {
		return errors.Wrap(err, "creating access service")
	}

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
	ts, err := timer.New(db)
	if err != nil {
		return errors.Wrap(err, "creating timer service")
	}

	for id, expiry := range s.Timers {
		if err := ts.SetTimer(id, expiry.Time); err != nil {
			return errors.Wrap(err, "storing counter in database")
		}
	}

	return nil
}

func (s storageFile) migrateVariables(db database.Connector) (err error) {
	for key, value := range s.Variables {
		if err := variables.SetVariable(db, key, value); err != nil {
			return errors.Wrap(err, "updating value in database")
		}
	}

	return nil
}

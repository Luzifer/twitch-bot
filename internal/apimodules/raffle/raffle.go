// Package raffle contains the backend and API implementation as well
// as the chat listeners for chat-raffles
package raffle

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v3/pkg/database"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "raffle"

var (
	db            database.Connector
	dbc           *dbClient
	formatMessage plugins.MsgFormatter
	send          plugins.SendMessageFunc
	tcGetter      func(string) (*twitch.Client, error)
)

func Register(args plugins.RegistrationArguments) error {
	db = args.GetDatabaseConnector()
	if err := db.DB().AutoMigrate(&raffle{}, &raffleEntry{}); err != nil {
		return errors.Wrap(err, "applying schema migration")
	}

	dbc = newDBClient(db)
	if err := dbc.RefreshActiveRaffles(); err != nil {
		return errors.Wrap(err, "refreshing active raffle cache")
	}

	formatMessage = args.FormatMessage
	send = args.SendMessage
	tcGetter = args.GetTwitchClientForChannel

	// FIXME: API routes

	if _, err := args.RegisterCron("@every 1m", func() {
		for name, fn := range map[string]func() error{
			"close":          dbc.AutoCloseExpired,
			"start":          dbc.AutoStart,
			"send_reminders": dbc.AutoSendReminders,
		} {
			if err := fn(); err != nil {
				logrus.WithFields(logrus.Fields{
					"actor": actorName,
					"cron":  name,
				}).WithError(err).Error("executing cron action")
			}
		}
	}); err != nil {
		return errors.Wrap(err, "registering cron")
	}

	if err := args.RegisterRawMessageHandler(rawMessageHandler); err != nil {
		return errors.Wrap(err, "registering raw message handler")
	}

	return nil
}

// Package raffle contains the backend and API implementation as well
// as the chat listeners for chat-raffles
package raffle

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/Luzifer/twitch-bot/v3/pkg/database"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

var (
	db             database.Connector
	dbc            *dbClient
	formatMessage  plugins.MsgFormatter
	frontendNotify func(string)
	send           plugins.SendMessageFunc
	tcGetter       func(string) (*twitch.Client, error)
)

func Register(args plugins.RegistrationArguments) (err error) {
	db = args.GetDatabaseConnector()
	if err := db.DB().AutoMigrate(&raffle{}, &raffleEntry{}); err != nil {
		return errors.Wrap(err, "applying schema migration")
	}

	args.RegisterCopyDatabaseFunc("raffle", func(src, target *gorm.DB) error {
		return database.CopyObjects(src, target, &raffle{}, &raffleEntry{})
	})

	dbc = newDBClient(db)
	if err = dbc.RefreshActiveRaffles(); err != nil {
		return errors.Wrap(err, "refreshing active raffle cache")
	}
	if err = dbc.RefreshSpeakUp(); err != nil {
		return errors.Wrap(err, "refreshing active speak-ups")
	}

	formatMessage = args.FormatMessage
	frontendNotify = args.FrontendNotify
	send = args.SendMessage
	tcGetter = args.GetTwitchClientForChannel

	if err = registerAPI(args); err != nil {
		return errors.Wrap(err, "registering API")
	}

	if _, err := args.RegisterCron("@every 10s", func() {
		for name, fn := range map[string]func() error{
			"close":          dbc.AutoCloseExpired,
			"start":          dbc.AutoStart,
			"send_reminders": dbc.AutoSendReminders,
		} {
			if err := fn(); err != nil {
				logrus.WithFields(logrus.Fields{
					"actor": moduleName,
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

	args.RegisterActor(enterRaffleActor{}.Name(), func() plugins.Actor { return &enterRaffleActor{} })
	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Enter user to raffle through channelpoints",
		Name:        "Enter User to Raffle",
		Type:        enterRaffleActor{}.Name(),

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "The keyword for the active raffle to enter the user into",
				Key:             "keyword",
				Name:            "Keyword",
				Optional:        false,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	return nil
}

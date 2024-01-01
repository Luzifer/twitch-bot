// Package log contains an actor to write bot-log entries from a rule
package log

import (
	"github.com/pkg/errors"
	"gopkg.in/irc.v4"

	"github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

var (
	formatMessage  plugins.MsgFormatter
	ptrStringEmpty = func(v string) *string { return &v }("")
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	formatMessage = args.FormatMessage

	args.RegisterActor("log", func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Print info log-line to bot log",
		Name:        "Log output",
		Type:        "log",

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Messsage to log into bot-log",
				Key:             "message",
				Name:            "Message",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	return nil
}

type actor struct{}

func (actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	message, err := formatMessage(attrs.MustString("message", ptrStringEmpty), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "executing message template")
	}

	logrus.WithFields(logrus.Fields{
		"channel":  plugins.DeriveChannel(m, eventData),
		"rule":     r.UUID,
		"username": plugins.DeriveUser(m, eventData),
	}).Info(message)
	return false, nil
}

func (actor) IsAsync() bool { return true }
func (actor) Name() string  { return "log" }

func (actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *plugins.FieldCollection) (err error) {
	if v, err := attrs.String("message"); err != nil || v == "" {
		return errors.New("message must be non-empty string")
	}

	if err = tplValidator(attrs.MustString("message", ptrStringEmpty)); err != nil {
		return errors.Wrap(err, "validating message template")
	}

	return nil
}

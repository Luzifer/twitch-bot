// Package logactor contains an actor to write bot-log entries from a rule
package logactor

import (
	"fmt"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"github.com/sirupsen/logrus"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

type actor struct{}

var formatMessage plugins.MsgFormatter

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

func (actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	message, err := formatMessage(attrs.MustString("message", helpers.Ptr("")), m, r, eventData)
	if err != nil {
		return false, fmt.Errorf("executing message template: %w", err)
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

func (actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "message", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.MustHaveNoUnknowFields,
		helpers.SchemaValidateTemplateField(tplValidator, "message"),
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	return nil
}

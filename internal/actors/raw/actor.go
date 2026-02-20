// Package raw contains an actor to send raw IRC messages
package raw

import (
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "raw"

var (
	formatMessage plugins.MsgFormatter
	send          plugins.SendMessageFunc
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	formatMessage = args.FormatMessage
	send = args.SendMessage

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Send raw IRC message",
		Name:        "Send RAW Message",
		Type:        "raw",

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Raw message to send (must be a valid IRC protocol message)",
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

func (actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	rawMsg, err := formatMessage(attrs.MustString("message", nil), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "preparing raw message")
	}

	msg, err := irc.ParseMessage(rawMsg)
	if err != nil {
		return false, errors.Wrap(err, "parsing raw message")
	}

	return false, errors.Wrap(
		send(msg),
		"sending raw message",
	)
}

func (actor) IsAsync() bool { return false }
func (actor) Name() string  { return actorName }

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

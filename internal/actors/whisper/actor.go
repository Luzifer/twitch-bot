// Package whisper contains an actor to send whispers
package whisper

import (
	"context"
	"fmt"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "whisper"

type actor struct{}

var (
	botTwitchClient func() *twitch.Client
	formatMessage   plugins.MsgFormatter
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	botTwitchClient = args.GetTwitchClient
	formatMessage = args.FormatMessage

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Send a whisper",
		Name:        "Send Whisper",
		Type:        "whisper",

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Message to whisper to the user",
				Key:             "message",
				Name:            "Message",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "",
				Description:     "User to send the message to",
				Key:             "to",
				Name:            "To User",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	return nil
}

func (actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	to, err := formatMessage(attrs.MustString("to", nil), m, r, eventData)
	if err != nil {
		return false, fmt.Errorf("preparing whisper receiver: %w", err)
	}

	msg, err := formatMessage(attrs.MustString("message", nil), m, r, eventData)
	if err != nil {
		return false, fmt.Errorf("preparing whisper message: %w", err)
	}

	if err = botTwitchClient().SendWhisper(context.Background(), to, msg); err != nil {
		return false, fmt.Errorf("sending whisper: %w", err)
	}

	return false, nil
}

func (actor) IsAsync() bool { return false }
func (actor) Name() string  { return actorName }

func (actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "message", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "to", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.MustHaveNoUnknowFields,
		helpers.SchemaValidateTemplateField(tplValidator, "message", "to"),
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	return nil
}

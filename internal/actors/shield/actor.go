// Package shield contains an actor to update the shield-mode for a
// given channel
package shield

import (
	"context"
	"fmt"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "shield"

type actor struct{}

var botTwitchClient func() *twitch.Client

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	botTwitchClient = args.GetTwitchClient

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Update shield mode for the given channel",
		Name:        "Update Shield Mode",
		Type:        actorName,

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "false",
				Description:     "Whether the shield-mode should be enabled or disabled",
				Key:             "enable",
				Name:            "Enable",
				Optional:        false,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeBool,
			},
		},
	})

	return nil
}

func (actor) Execute(_ *irc.Client, m *irc.Message, _ *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	if err = botTwitchClient().UpdateShieldMode(
		context.Background(),
		plugins.DeriveChannel(m, eventData),
		attrs.MustBool("enable", helpers.Ptr(false)),
	); err != nil {
		return false, fmt.Errorf("configuring shield mode: %w", err)
	}

	return false, nil
}

func (actor) IsAsync() bool { return false }
func (actor) Name() string  { return actorName }

func (actor) Validate(_ plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "enable", Type: fieldcollection.SchemaFieldTypeBool}),
		fieldcollection.MustHaveNoUnknowFields,
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	return nil
}

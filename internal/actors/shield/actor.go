package shield

import (
	"context"

	"github.com/pkg/errors"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "shield"

var botTwitchClient *twitch.Client

func Register(args plugins.RegistrationArguments) error {
	botTwitchClient = args.GetTwitchClient()

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

type actor struct{}

func (a actor) Execute(_ *irc.Client, m *irc.Message, _ *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	ptrBoolFalse := func(v bool) *bool { return &v }(false)

	return false, errors.Wrap(
		botTwitchClient.UpdateShieldMode(
			context.Background(),
			plugins.DeriveChannel(m, eventData),
			attrs.MustBool("enable", ptrBoolFalse),
		),
		"configuring shield mode",
	)
}

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return actorName }

func (a actor) Validate(_ plugins.TemplateValidatorFunc, attrs *plugins.FieldCollection) (err error) {
	if _, err = attrs.Bool("enable"); err != nil {
		return errors.New("enable must be boolean")
	}

	return nil
}

package deleteactor

import (
	"github.com/go-irc/irc"
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "delete"

var botTwitchClient *twitch.Client

func Register(args plugins.RegistrationArguments) error {
	botTwitchClient = args.GetTwitchClient()

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Delete message which caused the rule to be executed",
		Name:        "Delete Message",
		Type:        "delete",
	})

	return nil
}

type actor struct{}

func (a actor) Execute(_ *irc.Client, m *irc.Message, _ *plugins.Rule, eventData *plugins.FieldCollection, _ *plugins.FieldCollection) (preventCooldown bool, err error) {
	msgID, ok := m.Tags.GetTag("id")
	if !ok || msgID == "" {
		return false, nil
	}

	return false, errors.Wrap(
		botTwitchClient.DeleteMessage(
			plugins.DeriveChannel(m, eventData),
			msgID,
		),
		"deleting message",
	)
}

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return actorName }

func (a actor) Validate(plugins.TemplateValidatorFunc, *plugins.FieldCollection) (err error) {
	return nil
}

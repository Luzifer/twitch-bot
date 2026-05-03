// Package deleteactor contains an actor to delete messages
package deleteactor

import (
	"context"
	"fmt"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "delete"

type actor struct{}

var botTwitchClient func() *twitch.Client

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	botTwitchClient = args.GetTwitchClient

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Delete message which caused the rule to be executed",
		Name:        "Delete Message",
		Type:        "delete",
	})

	return nil
}

func (actor) Execute(_ *irc.Client, m *irc.Message, _ *plugins.Rule, eventData *fieldcollection.FieldCollection, _ *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	msgID, ok := m.Tags["id"]
	if !ok || msgID == "" {
		return false, nil
	}

	if err = botTwitchClient().DeleteMessage(
		context.Background(),
		plugins.DeriveChannel(m, eventData),
		msgID,
	); err != nil {
		return false, fmt.Errorf("deleting message: %w", err)
	}

	return false, nil
}

func (actor) IsAsync() bool { return false }
func (actor) Name() string  { return actorName }

func (actor) Validate(plugins.TemplateValidatorFunc, *fieldcollection.FieldCollection) (err error) {
	return nil
}

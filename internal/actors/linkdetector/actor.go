package linkdetector

import (
	"github.com/go-irc/irc"

	"github.com/Luzifer/twitch-bot/v3/internal/linkcheck"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "linkdetector"

func Register(args plugins.RegistrationArguments) error {
	args.RegisterActor(actorName, func() plugins.Actor { return &Actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: `Scans for links in the message and adds the "links" field to the event data`,
		Name:        "Scan for Links",
		Type:        actorName,
	})

	return nil
}

type Actor struct{}

func (Actor) Execute(_ *irc.Client, m *irc.Message, _ *plugins.Rule, eventData *plugins.FieldCollection, _ *plugins.FieldCollection) (preventCooldown bool, err error) {
	if eventData.HasAll("links") {
		// We already detected links, lets not do it again
		return false, nil
	}

	eventData.Set("links", linkcheck.New().ScanForLinks(m.Trailing()))
	return false, nil
}

func (Actor) IsAsync() bool { return false }

func (Actor) Name() string { return actorName }

func (Actor) Validate(plugins.TemplateValidatorFunc, *plugins.FieldCollection) error { return nil }

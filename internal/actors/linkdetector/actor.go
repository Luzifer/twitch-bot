package linkdetector

import (
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/internal/linkcheck"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "linkdetector"

var ptrFalse = func(v bool) *bool { return &v }(false)

func Register(args plugins.RegistrationArguments) error {
	args.RegisterActor(actorName, func() plugins.Actor { return &Actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: `Scans for links in the message and adds the "links" field to the event data`,
		Name:        "Scan for Links",
		Type:        actorName,

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "false",
				Description:     "Enable heuristic scans to find links with spaces or other means of obfuscation in them",
				Key:             "heuristic",
				Name:            "Heuristic Scan",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeBool,
			},
		},
	})

	return nil
}

type Actor struct{}

func (Actor) Execute(_ *irc.Client, m *irc.Message, _ *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	if eventData.HasAll("links") {
		// We already detected links, lets not do it again
		return false, nil
	}

	if attrs.MustBool("heuristic", ptrFalse) {
		eventData.Set("links", linkcheck.New().HeuristicScanForLinks(m.Trailing()))
	} else {
		eventData.Set("links", linkcheck.New().ScanForLinks(m.Trailing()))
	}

	return false, nil
}

func (Actor) IsAsync() bool { return false }

func (Actor) Name() string { return actorName }

func (Actor) Validate(plugins.TemplateValidatorFunc, *plugins.FieldCollection) error { return nil }

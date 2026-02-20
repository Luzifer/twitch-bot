// Package linkdetector contains an actor to detect links in a message
// and add them to a variable
package linkdetector

import (
	"fmt"

	"gopkg.in/irc.v4"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/internal/linkcheck"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "linkdetector"

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	args.RegisterActor(actorName, func() plugins.Actor { return &Actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: `Scans for links in the message and adds the "links" field to the event data`,
		Name:        "Scan for Links",
		Type:        actorName,

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "false",
				Description:     "Enable heuristic scans to find links with spaces or other means of obfuscation in them (quite slow and will detect MANY false-positive links, only use for blacklisting links!)",
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

// Actor implements the actor interface
type Actor struct{}

// Execute implements the actor interface
func (Actor) Execute(_ *irc.Client, m *irc.Message, _ *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	if eventData.HasAll("links") {
		// We already detected links, lets not do it again
		return false, nil
	}

	if attrs.MustBool("heuristic", helpers.Ptr(false)) {
		eventData.Set("links", linkcheck.New().HeuristicScanForLinks(m.Trailing()))
	} else {
		eventData.Set("links", linkcheck.New().ScanForLinks(m.Trailing()))
	}

	return false, nil
}

// IsAsync implements the actor interface
func (Actor) IsAsync() bool { return false }

// Name implements the actor interface
func (Actor) Name() string { return actorName }

// Validate implements the actor interface
func (Actor) Validate(_ plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "heuristic", Type: fieldcollection.SchemaFieldTypeBool}),
		fieldcollection.MustHaveNoUnknowFields,
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	return nil
}

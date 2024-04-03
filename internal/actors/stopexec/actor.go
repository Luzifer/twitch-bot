// Package stopexec contains an actor to stop the rule execution on
// template condition
package stopexec

import (
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/go_helpers/v2/fieldcollection"
	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "stopexec"

var formatMessage plugins.MsgFormatter

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	formatMessage = args.FormatMessage

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Stop Rule Execution on Condition",
		Name:        "Stop Execution",
		Type:        actorName,

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Condition when to stop execution (must evaluate to \"true\" to stop execution)",
				Key:             "when",
				Name:            "When",
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
	when, err := formatMessage(attrs.MustString("when", helpers.Ptr("")), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "executing when template")
	}

	if when == "true" {
		return false, plugins.ErrStopRuleExecution
	}

	return false, nil
}

func (actor) IsAsync() bool { return false }
func (actor) Name() string  { return actorName }

func (actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "when", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		helpers.SchemaValidateTemplateField(tplValidator, "when"),
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	return nil
}

// Package eventmod contains an actor to modify event data during rule
// execution by adding fields (template variables)
package eventmod

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "eventmod"

type actor struct{}

var formatMessage plugins.MsgFormatter

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	formatMessage = args.FormatMessage

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Add custom fields to the event to be used as template variables later on",
		Name:        "Add Fields to Event",
		Type:        actorName,

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Fields to set in the event (must produce valid JSON: `map[string]any`)",
				Key:             "fields",
				Name:            "Fields",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	return nil
}

func (actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	ptrStringEmpty := func(v string) *string { return &v }("")

	fd, err := formatMessage(attrs.MustString("fields", ptrStringEmpty), m, r, eventData)
	if err != nil {
		return false, fmt.Errorf("executing fields template: %w", err)
	}

	if fd == "" {
		return false, errors.New("fields template evaluated to empty string")
	}

	fields := make(map[string]any)
	if err = json.Unmarshal([]byte(fd), &fields); err != nil {
		return false, fmt.Errorf("parsing fields: %w", err)
	}

	eventData.SetFromData(fields)

	return false, nil
}

func (actor) IsAsync() bool { return false }
func (actor) Name() string  { return actorName }

func (actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "fields", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.MustHaveNoUnknowFields,
		helpers.SchemaValidateTemplateField(tplValidator, "fields"),
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	return nil
}

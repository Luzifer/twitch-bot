package eventmod

import (
	"encoding/json"

	"github.com/pkg/errors"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "eventmod"

var formatMessage plugins.MsgFormatter

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

type actor struct{}

func (a actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	ptrStringEmpty := func(v string) *string { return &v }("")

	fd, err := formatMessage(attrs.MustString("fields", ptrStringEmpty), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "executing fields template")
	}

	if fd == "" {
		return false, errors.New("fields template evaluated to empty string")
	}

	fields := map[string]any{}
	if err = json.Unmarshal([]byte(fd), &fields); err != nil {
		return false, errors.Wrap(err, "parsing fields")
	}

	eventData.SetFromData(fields)

	return false, nil
}

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return actorName }

func (a actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *plugins.FieldCollection) (err error) {
	fieldsTemplate, err := attrs.String("fields")
	if err != nil || fieldsTemplate == "" {
		return errors.New("fields must be non-empty string")
	}

	if err = tplValidator(fieldsTemplate); err != nil {
		return errors.Wrap(err, "validating fields template")
	}

	return nil
}

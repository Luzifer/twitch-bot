package stopexec

import (
	"github.com/pkg/errors"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "stopexec"

var formatMessage plugins.MsgFormatter

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

func (a actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	ptrStringEmpty := func(v string) *string { return &v }("")

	when, err := formatMessage(attrs.MustString("when", ptrStringEmpty), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "executing when template")
	}

	if when == "true" {
		return false, plugins.ErrStopRuleExecution
	}

	return false, nil
}

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return actorName }

func (a actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *plugins.FieldCollection) (err error) {
	whenTemplate, err := attrs.String("when")
	if err != nil || whenTemplate == "" {
		return errors.New("when must be non-empty string")
	}

	if err = tplValidator(whenTemplate); err != nil {
		return errors.Wrap(err, "validating when template")
	}

	return nil
}

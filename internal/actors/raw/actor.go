// Package raw contains an actor to send raw IRC messages
package raw

import (
	"github.com/pkg/errors"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "raw"

var (
	formatMessage plugins.MsgFormatter
	send          plugins.SendMessageFunc

	ptrStringEmpty = func(s string) *string { return &s }("")
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	formatMessage = args.FormatMessage
	send = args.SendMessage

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Send raw IRC message",
		Name:        "Send RAW Message",
		Type:        "raw",

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Raw message to send (must be a valid IRC protocol message)",
				Key:             "message",
				Name:            "Message",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	return nil
}

type actor struct{}

func (actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	rawMsg, err := formatMessage(attrs.MustString("message", nil), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "preparing raw message")
	}

	msg, err := irc.ParseMessage(rawMsg)
	if err != nil {
		return false, errors.Wrap(err, "parsing raw message")
	}

	return false, errors.Wrap(
		send(msg),
		"sending raw message",
	)
}

func (actor) IsAsync() bool { return false }
func (actor) Name() string  { return actorName }

func (actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *plugins.FieldCollection) (err error) {
	if v, err := attrs.String("message"); err != nil || v == "" {
		return errors.New("message must be non-empty string")
	}

	if err = tplValidator(attrs.MustString("message", ptrStringEmpty)); err != nil {
		return errors.Wrap(err, "validating message template")
	}

	return nil
}

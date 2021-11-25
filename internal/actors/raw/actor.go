package raw

import (
	"github.com/go-irc/irc"
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/plugins"
)

const actorName = "raw"

var formatMessage plugins.MsgFormatter

func Register(args plugins.RegistrationArguments) error {
	formatMessage = args.FormatMessage

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

func (a actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	rawMsg, err := formatMessage(attrs.MustString("message", nil), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "preparing raw message")
	}

	msg, err := irc.ParseMessage(rawMsg)
	if err != nil {
		return false, errors.Wrap(err, "parsing raw message")
	}

	return false, errors.Wrap(
		c.WriteMessage(msg),
		"sending raw message",
	)
}

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return actorName }

func (a actor) Validate(attrs *plugins.FieldCollection) (err error) {
	if v, err := attrs.String("message"); err != nil || v == "" {
		return errors.New("message must be non-empty string")
	}

	return nil
}

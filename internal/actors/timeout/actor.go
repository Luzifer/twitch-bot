package timeout

import (
	"fmt"
	"time"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

const actorName = "timeout"

func Register(args plugins.RegistrationArguments) error {
	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Timeout user from chat",
		Name:        "Timeout",
		Type:        "timeout",

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Duration of the timeout",
				Key:             "duration",
				Name:            "Duration",
				Optional:        false,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeDuration,
			},
		},
	})

	return nil
}

type actor struct{}

func (a actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData plugins.FieldCollection, attrs plugins.FieldCollection) (preventCooldown bool, err error) {
	return false, errors.Wrap(
		c.WriteMessage(&irc.Message{
			Command: "PRIVMSG",
			Params: []string{
				plugins.DeriveChannel(m, eventData),
				fmt.Sprintf("/timeout %s %d", plugins.DeriveUser(m, eventData), attrs.MustDuration("duration", nil)/time.Second),
			},
		}),
		"sending timeout",
	)
}

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return actorName }

func (a actor) Validate(attrs plugins.FieldCollection) (err error) {
	if v, err := attrs.Duration("duration"); err != nil || v >= time.Second {
		return errors.New("duration must be of type duration greater or equal one second")
	}

	return nil
}

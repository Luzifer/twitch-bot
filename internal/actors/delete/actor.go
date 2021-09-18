package deleteactor

import (
	"fmt"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

const actorName = "delete"

func Register(args plugins.RegistrationArguments) error {
	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	return nil
}

type actor struct {
	DeleteMessage *bool `json:"delete_message" yaml:"delete_message"`
}

func (a actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData plugins.FieldCollection) (preventCooldown bool, err error) {
	if a.DeleteMessage == nil || !*a.DeleteMessage || m == nil {
		return false, nil
	}

	msgID, ok := m.Tags.GetTag("id")
	if !ok || msgID == "" {
		return false, nil
	}

	return false, errors.Wrap(
		c.WriteMessage(&irc.Message{
			Command: "PRIVMSG",
			Params: []string{
				m.Params[0],
				fmt.Sprintf("/delete %s", msgID),
			},
		}),
		"sending delete",
	)
}

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return actorName }

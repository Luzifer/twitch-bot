package deleteActor

import (
	"fmt"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func Register(args plugins.RegistrationArguments) error {
	args.RegisterActor(func() plugins.Actor { return &ActorDelete{} })

	return nil
}

type ActorDelete struct {
	DeleteMessage *bool `json:"delete_message" yaml:"delete_message"`
}

func (a ActorDelete) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule) (preventCooldown bool, err error) {
	if a.DeleteMessage == nil || !*a.DeleteMessage {
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

func (a ActorDelete) IsAsync() bool { return false }
func (a ActorDelete) Name() string  { return "delete" }

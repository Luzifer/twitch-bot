package ban

import (
	"fmt"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func Register(args plugins.RegistrationArguments) error {
	args.RegisterActor(func() plugins.Actor { return &actor{} })

	return nil
}

type actor struct {
	Ban *string `json:"ban" yaml:"ban"`
}

func (a actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule) (preventCooldown bool, err error) {
	if a.Ban == nil {
		return false, nil
	}

	return false, errors.Wrap(
		c.WriteMessage(&irc.Message{
			Command: "PRIVMSG",
			Params: []string{
				m.Params[0],
				fmt.Sprintf("/ban %s %s", m.User, *a.Ban),
			},
		}),
		"sending timeout",
	)
}

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return "ban" }

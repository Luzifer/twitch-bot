package main

import (
	"fmt"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/plugins"
)

var _ plugins.RegisterFunc = Register

func Register(args plugins.RegistrationArguments) error {
	args.GetLogger("plugin-example").Warn("Example Register called")

	args.RegisterActor(func() plugins.Actor { return &actor{} })

	return nil
}

type actor struct {
	Example bool `json:"example" yaml:"example"`
}

func (a actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule) (preventCooldown bool, err error) {
	if !a.Example {
		return false, nil
	}

	return false, errors.Wrap(
		c.WriteMessage(&irc.Message{
			Command: "PRIVMSG",
			Params: []string{
				m.Params[0],
				fmt.Sprintf("@%s Example plugin noticed you! KonCha", m.User),
			},
		}),
		"sending response",
	)
}

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return "example" }

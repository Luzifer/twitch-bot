package main

import (
	"fmt"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func init() {
	registerAction(func() Actor { return &ActorBan{} })
}

type ActorBan struct {
	Ban *string `json:"ban" yaml:"ban"`
}

func (a ActorBan) Execute(c *irc.Client, m *irc.Message, r *Rule) error {
	if a.Ban == nil {
		return nil
	}

	return errors.Wrap(
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

func (a ActorBan) IsAsync() bool { return false }
func (a ActorBan) Name() string  { return "ban" }

package main

import (
	"fmt"
	"time"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func init() {
	registerAction(func() Actor { return &ActorTimeout{} })
}

type ActorTimeout struct {
	Timeout *time.Duration `json:"timeout" yaml:"timeout"`
}

func (a ActorTimeout) Execute(c *irc.Client, m *irc.Message, r *Rule) (preventCooldown bool, err error) {
	if a.Timeout == nil {
		return false, nil
	}

	return false, errors.Wrap(
		c.WriteMessage(&irc.Message{
			Command: "PRIVMSG",
			Params: []string{
				m.Params[0],
				fmt.Sprintf("/timeout %s %d", m.User, fixDurationValue(*a.Timeout)/time.Second),
			},
		}),
		"sending timeout",
	)
}

func (a ActorTimeout) IsAsync() bool { return false }
func (a ActorTimeout) Name() string  { return "timeout" }

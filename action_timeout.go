package main

import (
	"fmt"
	"time"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func init() {
	registerAction(func(c *irc.Client, m *irc.Message, ruleDef *Rule, r *RuleAction) error {
		if r.Timeout == nil {
			return nil
		}

		return errors.Wrap(
			c.WriteMessage(&irc.Message{
				Command: "PRIVMSG",
				Params: []string{
					m.Params[0],
					fmt.Sprintf("/timeout %s %d", m.User, *r.Timeout/time.Second),
				},
			}),
			"sending timeout",
		)
	})
}

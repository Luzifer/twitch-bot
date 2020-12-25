package main

import (
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func init() {
	registerAction(func(c *irc.Client, m *irc.Message, ruleDef *rule, r *ruleAction) error {
		if r.Respond == nil {
			return nil
		}

		msg, err := formatMessage(*r.Respond, m, ruleDef, nil)
		if err != nil {
			return errors.Wrap(err, "preparing response")
		}

		return errors.Wrap(
			c.WriteMessage(&irc.Message{
				Command: "PRIVMSG",
				Params: []string{
					m.Params[0],
					msg,
				},
			}),
			"sending response",
		)
	})
}

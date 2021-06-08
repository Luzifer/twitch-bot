package main

import (
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func init() {
	registerAction(func(c *irc.Client, m *irc.Message, ruleDef *Rule, r *RuleAction) error {
		if r.RawMessage == nil {
			return nil
		}

		rawMsg, err := formatMessage(*r.RawMessage, m, ruleDef, nil)
		if err != nil {
			return errors.Wrap(err, "preparing raw message")
		}

		msg, err := irc.ParseMessage(rawMsg)
		if err != nil {
			return errors.Wrap(err, "parsing raw message")
		}

		return errors.Wrap(
			c.WriteMessage(msg),
			"sending raw message",
		)
	})
}

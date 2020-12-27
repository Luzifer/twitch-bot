package main

import (
	"fmt"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func init() {
	registerAction(func(c *irc.Client, m *irc.Message, ruleDef *rule, r *ruleAction) error {
		if r.DeleteMessage == nil || !*r.DeleteMessage {
			return nil
		}

		msgID, ok := m.Tags.GetTag("id")
		if !ok || msgID == "" {
			return nil
		}

		return errors.Wrap(
			c.WriteMessage(&irc.Message{
				Command: "PRIVMSG",
				Params: []string{
					m.Params[0],
					fmt.Sprintf("/delete %s", msgID),
				},
			}),
			"sending delete",
		)
	})
}

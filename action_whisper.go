package main

import (
	"fmt"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func init() {
	registerAction(func(c *irc.Client, m *irc.Message, ruleDef *Rule, r *RuleAction) error {
		if r.WhisperTo == nil || r.WhisperMessage == nil {
			return nil
		}

		to, err := formatMessage(*r.WhisperTo, m, ruleDef, nil)
		if err != nil {
			return errors.Wrap(err, "preparing whisper receiver")
		}

		msg, err := formatMessage(*r.WhisperMessage, m, ruleDef, nil)
		if err != nil {
			return errors.Wrap(err, "preparing whisper message")
		}

		channel := "#tmijs" // As a fallback, copied from tmi.js
		if len(config.Channels) > 0 {
			channel = fmt.Sprintf("#%s", config.Channels[0])
		}

		return errors.Wrap(
			c.WriteMessage(&irc.Message{
				Command: "PRIVMSG",
				Params: []string{
					channel,
					fmt.Sprintf("/w %s %s", to, msg),
				},
			}),
			"sending whisper",
		)
	})
}

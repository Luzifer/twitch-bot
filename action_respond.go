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
			if r.RespondFallback == nil {
				return errors.Wrap(err, "preparing response")
			}
			if msg, err = formatMessage(*r.RespondFallback, m, ruleDef, nil); err != nil {
				return errors.Wrap(err, "preparing response fallback")
			}
		}

		ircMessage := &irc.Message{
			Command: "PRIVMSG",
			Params: []string{
				m.Params[0],
				msg,
			},
		}

		if r.RespondAsReply != nil && *r.RespondAsReply {
			id, ok := m.GetTag("id")
			if ok {
				if ircMessage.Tags == nil {
					ircMessage.Tags = make(irc.Tags)
				}
				ircMessage.Tags["reply-parent-msg-id"] = irc.TagValue(id)
			}
		}

		return errors.Wrap(
			c.WriteMessage(ircMessage),
			"sending response",
		)
	})
}

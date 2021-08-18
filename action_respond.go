package main

import (
	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func init() {
	registerAction(func() plugins.Actor { return &ActorRespond{} })
}

type ActorRespond struct {
	Respond         *string `json:"respond" yaml:"respond"`
	RespondAsReply  *bool   `json:"respond_as_reply" yaml:"respond_as_reply"`
	RespondFallback *string `json:"respond_fallback" yaml:"respond_fallback"`
}

func (a ActorRespond) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule) (preventCooldown bool, err error) {
	if a.Respond == nil {
		return false, nil
	}

	msg, err := formatMessage(*a.Respond, m, r, nil)
	if err != nil {
		if a.RespondFallback == nil {
			return false, errors.Wrap(err, "preparing response")
		}
		if msg, err = formatMessage(*a.RespondFallback, m, r, nil); err != nil {
			return false, errors.Wrap(err, "preparing response fallback")
		}
	}

	ircMessage := &irc.Message{
		Command: "PRIVMSG",
		Params: []string{
			m.Params[0],
			msg,
		},
	}

	if a.RespondAsReply != nil && *a.RespondAsReply {
		id, ok := m.GetTag("id")
		if ok {
			if ircMessage.Tags == nil {
				ircMessage.Tags = make(irc.Tags)
			}
			ircMessage.Tags["reply-parent-msg-id"] = irc.TagValue(id)
		}
	}

	return false, errors.Wrap(
		c.WriteMessage(ircMessage),
		"sending response",
	)
}

func (a ActorRespond) IsAsync() bool { return false }
func (a ActorRespond) Name() string  { return "respond" }

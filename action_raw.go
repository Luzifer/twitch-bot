package main

import (
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func init() {
	registerAction(func() Actor { return &ActorRaw{} })
}

type ActorRaw struct {
	RawMessage *string `json:"raw_message" yaml:"raw_message"`
}

func (a ActorRaw) Execute(c *irc.Client, m *irc.Message, r *Rule) (preventCooldown bool, err error) {
	if a.RawMessage == nil {
		return false, nil
	}

	rawMsg, err := formatMessage(*a.RawMessage, m, r, nil)
	if err != nil {
		return false, errors.Wrap(err, "preparing raw message")
	}

	msg, err := irc.ParseMessage(rawMsg)
	if err != nil {
		return false, errors.Wrap(err, "parsing raw message")
	}

	return false, errors.Wrap(
		c.WriteMessage(msg),
		"sending raw message",
	)
}

func (a ActorRaw) IsAsync() bool { return false }
func (a ActorRaw) Name() string  { return "raw" }

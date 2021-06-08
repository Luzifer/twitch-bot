package main

import (
	"fmt"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func init() {
	registerAction(func() Actor { return &ActorDelete{} })
}

type ActorDelete struct {
	DeleteMessage *bool `json:"delete_message" yaml:"delete_message"`
}

func (a ActorDelete) Execute(c *irc.Client, m *irc.Message, r *Rule) error {
	if a.DeleteMessage == nil || !*a.DeleteMessage {
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
}

func (a ActorDelete) IsAsync() bool { return false }
func (a ActorDelete) Name() string  { return "delete" }

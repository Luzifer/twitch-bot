package main

import (
	"fmt"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func init() {
	registerAction(func() plugins.Actor { return &ActorWhisper{} })
}

type ActorWhisper struct {
	WhisperMessage *string `json:"whisper_message" yaml:"whisper_message"`
	WhisperTo      *string `json:"whisper_to" yaml:"whisper_to"`
}

func (a ActorWhisper) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule) (preventCooldown bool, err error) {
	if a.WhisperTo == nil || a.WhisperMessage == nil {
		return false, nil
	}

	to, err := formatMessage(*a.WhisperTo, m, r, nil)
	if err != nil {
		return false, errors.Wrap(err, "preparing whisper receiver")
	}

	msg, err := formatMessage(*a.WhisperMessage, m, r, nil)
	if err != nil {
		return false, errors.Wrap(err, "preparing whisper message")
	}

	channel := "#tmijs" // As a fallback, copied from tmi.js
	if len(config.Channels) > 0 {
		channel = fmt.Sprintf("#%s", config.Channels[0])
	}

	return false, errors.Wrap(
		c.WriteMessage(&irc.Message{
			Command: "PRIVMSG",
			Params: []string{
				channel,
				fmt.Sprintf("/w %s %s", to, msg),
			},
		}),
		"sending whisper",
	)
}

func (a ActorWhisper) IsAsync() bool { return false }
func (a ActorWhisper) Name() string  { return "whisper" }

package whisper

import (
	"fmt"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

const actorName = "whisper"

var formatMessage plugins.MsgFormatter

func Register(args plugins.RegistrationArguments) error {
	formatMessage = args.FormatMessage

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	return nil
}

type actor struct {
	WhisperMessage *string `json:"whisper_message" yaml:"whisper_message"`
	WhisperTo      *string `json:"whisper_to" yaml:"whisper_to"`
}

func (a actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData plugins.FieldCollection) (preventCooldown bool, err error) {
	if a.WhisperTo == nil || a.WhisperMessage == nil {
		return false, nil
	}

	to, err := formatMessage(*a.WhisperTo, m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "preparing whisper receiver")
	}

	msg, err := formatMessage(*a.WhisperMessage, m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "preparing whisper message")
	}

	channel := "#tmijs" // As a fallback, copied from tmi.js

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

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return actorName }

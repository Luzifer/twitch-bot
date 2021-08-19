package raw

import (
	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

var formatMessage plugins.MsgFormatter

func Register(args plugins.RegistrationArguments) error {
	formatMessage = args.FormatMessage

	args.RegisterActor(func() plugins.Actor { return &actor{} })

	return nil
}

type actor struct {
	RawMessage *string `json:"raw_message" yaml:"raw_message"`
}

func (a actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule) (preventCooldown bool, err error) {
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

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return "raw" }

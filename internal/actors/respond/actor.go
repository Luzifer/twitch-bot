package respond

import (
	"fmt"
	"strings"

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
	Respond         *string `json:"respond" yaml:"respond"`
	RespondAsReply  *bool   `json:"respond_as_reply" yaml:"respond_as_reply"`
	RespondFallback *string `json:"respond_fallback" yaml:"respond_fallback"`
	ToChannel       *string `json:"to_channel" yaml:"to_channel"`
}

func (a actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData plugins.FieldCollection) (preventCooldown bool, err error) {
	if a.Respond == nil {
		return false, nil
	}

	msg, err := formatMessage(*a.Respond, m, r, eventData)
	if err != nil {
		if a.RespondFallback == nil {
			return false, errors.Wrap(err, "preparing response")
		}
		if msg, err = formatMessage(*a.RespondFallback, m, r, eventData); err != nil {
			return false, errors.Wrap(err, "preparing response fallback")
		}
	}

	toChannel := plugins.DeriveChannel(m, eventData)
	if a.ToChannel != nil {
		toChannel = fmt.Sprintf("#%s", strings.TrimLeft(*a.ToChannel, "#"))
	}

	ircMessage := &irc.Message{
		Command: "PRIVMSG",
		Params: []string{
			toChannel,
			msg,
		},
	}

	if a.RespondAsReply != nil && *a.RespondAsReply && m != nil {
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

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return "respond" }

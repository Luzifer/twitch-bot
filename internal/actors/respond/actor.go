package respond

import (
	"fmt"
	"strings"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

const actorName = "respond"

var (
	formatMessage plugins.MsgFormatter

	ptrBoolFalse = func(v bool) *bool { return &v }(false)
)

func Register(args plugins.RegistrationArguments) error {
	formatMessage = args.FormatMessage

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	return nil
}

type actor struct{}

func (a actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData plugins.FieldCollection, attrs plugins.FieldCollection) (preventCooldown bool, err error) {
	msg, err := formatMessage(attrs.MustString("message", nil), m, r, eventData)
	if err != nil {
		if attrs.CanString("fallback") {
			return false, errors.Wrap(err, "preparing response")
		}
		if msg, err = formatMessage(attrs.MustString("fallback", nil), m, r, eventData); err != nil {
			return false, errors.Wrap(err, "preparing response fallback")
		}
	}

	toChannel := plugins.DeriveChannel(m, eventData)
	if attrs.CanString("to_channel") {
		toChannel = fmt.Sprintf("#%s", strings.TrimLeft(attrs.MustString("to_channel", nil), "#"))
	}

	ircMessage := &irc.Message{
		Command: "PRIVMSG",
		Params: []string{
			toChannel,
			msg,
		},
	}

	if attrs.MustBool("as_reply", ptrBoolFalse) {
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
func (a actor) Name() string  { return actorName }

func (a actor) Validate(attrs plugins.FieldCollection) (err error) {
	if v, err := attrs.String("message"); err != nil || v == "" {
		return errors.New("message must be non-empty string")
	}

	return nil
}

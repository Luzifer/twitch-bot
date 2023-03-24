package respond

import (
	"fmt"
	"strings"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "respond"

var (
	formatMessage plugins.MsgFormatter
	send          plugins.SendMessageFunc

	ptrBoolFalse   = func(v bool) *bool { return &v }(false)
	ptrStringEmpty = func(s string) *string { return &s }("")
)

func Register(args plugins.RegistrationArguments) error {
	formatMessage = args.FormatMessage
	send = args.SendMessage

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Respond to message with a new message",
		Name:        "Respond to Message",
		Type:        "respond",

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Message text to send",
				Key:             "message",
				Long:            true,
				Name:            "Message",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "",
				Description:     "Fallback message text to send if message cannot be generated",
				Key:             "fallback",
				Name:            "Fallback",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "false",
				Description:     "Send message as a native Twitch-reply to the original message",
				Key:             "as_reply",
				Name:            "As Reply",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeBool,
			},
			{
				Default:         "",
				Description:     "Send message to a different channel than the original message",
				Key:             "to_channel",
				Name:            "To Channel",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	return nil
}

type actor struct{}

func (a actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	msg, err := formatMessage(attrs.MustString("message", nil), m, r, eventData)
	if err != nil {
		if !attrs.CanString("fallback") || attrs.MustString("fallback", nil) == "" {
			return false, errors.Wrap(err, "preparing response")
		}
		log.WithError(err).Error("Response message processing caused error, trying fallback")
		if msg, err = formatMessage(attrs.MustString("fallback", nil), m, r, eventData); err != nil {
			return false, errors.Wrap(err, "preparing response fallback")
		}
	}

	toChannel := plugins.DeriveChannel(m, eventData)
	if attrs.CanString("to_channel") && attrs.MustString("to_channel", nil) != "" {
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
		send(ircMessage),
		"sending response",
	)
}

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return actorName }

func (a actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *plugins.FieldCollection) (err error) {
	if v, err := attrs.String("message"); err != nil || v == "" {
		return errors.New("message must be non-empty string")
	}

	for _, field := range []string{"message", "fallback"} {
		if err = tplValidator(attrs.MustString(field, ptrStringEmpty)); err != nil {
			return errors.Wrapf(err, "validating %s template", field)
		}
	}

	return nil
}

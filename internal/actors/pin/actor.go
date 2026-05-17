// Package pin contains an actor to send and pin / to pin an existing
// message to a given channel
package pin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const (
	actorName = "pin"

	minDuration = 30 * time.Second
	maxDuration = 1800 * time.Second
)

type actor struct{}

var (
	formatMessage plugins.MsgFormatter
	tcGetter      func() *twitch.Client
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	formatMessage = args.FormatMessage
	tcGetter = args.GetTwitchClient

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Pin a message to the channel",
		Name:        "Pin Message",
		Type:        actorName,

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Message to pin",
				Key:             "message",
				Long:            true,
				Name:            "Message",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "",
				Description:     "Message-ID to pin",
				Key:             "message_id",
				Name:            "Message ID",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "",
				Description:     `Duration of the pin (between 30s and 30m; empty for "until end of stream")`,
				Key:             "duration",
				Name:            "Duration",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeDuration,
			},
			{
				Default:         "",
				Description:     "Override channel to pin the message in",
				Key:             "channel",
				Name:            "Channel",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	return nil
}

func (a actor) Execute(
	_ *irc.Client,
	m *irc.Message,
	r *plugins.Rule,
	eventData *fieldcollection.FieldCollection,
	attrs *fieldcollection.FieldCollection,
) (preventCooldown bool, err error) {
	twitchClient := tcGetter()

	channel := plugins.DeriveChannel(m, eventData)
	if attrs.CanString("channel") && attrs.MustString("channel", nil) != "" {
		channel = fmt.Sprintf("#%s", strings.TrimLeft(attrs.MustString("channel", nil), "#"))
	}

	var msg, msgID string

	if attrs.MustString("message_id", new("")) != "" {
		if msgID, err = formatMessage(attrs.MustString("message_id", nil), m, r, eventData); err != nil {
			return false, fmt.Errorf("executing message_id template: %w", err)
		}
		msgID = strings.TrimSpace(msgID)
	}

	if attrs.MustString("message", new("")) != "" {
		if msg, err = formatMessage(attrs.MustString("message", nil), m, r, eventData); err != nil {
			return false, fmt.Errorf("executing message template: %w", err)
		}
		msg = strings.TrimSpace(msg)
	}

	switch {
	case msg != "" && msgID != "":
		return false, fmt.Errorf("message and message_id are both set non-empty")

	case msgID != "":
		return false, a.pinMessageIDToChannel(
			twitchClient,
			channel,
			msgID,
			attrs.MustDuration("duration", new(time.Duration(0))),
		)

	case msg != "":
		return false, a.pinMessageToChannel(
			twitchClient,
			channel,
			msg,
			attrs.MustDuration("duration", new(time.Duration(0))),
		)

	default:
		return false, fmt.Errorf("either message or message_id must be given")
	}
}

func (actor) IsAsync() bool { return false }
func (actor) Name() string  { return actorName }

func (actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "duration", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeDuration}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "channel", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "message", Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "message_id", Type: fieldcollection.SchemaFieldTypeString}),
		helpers.SchemaValidateTemplateField(tplValidator, "message", "message_id"),
		fieldcollection.MustHaveNoUnknowFields,
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	if d := attrs.MustDuration("duration", new(time.Duration(0))); d != 0 && (d < minDuration || d > maxDuration) {
		return fmt.Errorf("duration must be between 30s and 30m when set")
	}

	return nil
}

func (actor) pinMessageIDToChannel(
	tc *twitch.Client,
	channel, messageID string,
	duration time.Duration,
) (err error) {
	if err = tc.PinChatMessage(context.TODO(), channel, messageID, duration); err != nil {
		return fmt.Errorf("pinning message to channel: %w", err)
	}

	return nil
}

func (a actor) pinMessageToChannel(
	tc *twitch.Client,
	channel, message string,
	duration time.Duration,
) (err error) {
	res, err := tc.SendChatMessage(context.TODO(), channel, message, "", false, false)
	if err != nil {
		return fmt.Errorf("sending message: %w", err)
	}

	if !res.IsSent {
		var msg string
		if res.DropReason != nil {
			msg = res.DropReason.Message
		}
		return fmt.Errorf("message was not send because of: %s", msg)
	}

	return a.pinMessageIDToChannel(tc, channel, res.MessageID, duration)
}

// Package timeout contains an actor to timeout users
package timeout

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "timeout"

var (
	botTwitchClient func() *twitch.Client
	formatMessage   plugins.MsgFormatter
	ptrStringEmpty  = func(v string) *string { return &v }("")

	timeoutChatcommandRegex = regexp.MustCompile(`^/timeout +([^\s]+) +([0-9]+) +(.+)$`)
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	botTwitchClient = args.GetTwitchClient
	formatMessage = args.FormatMessage

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Timeout user from chat",
		Name:        "Timeout User",
		Type:        "timeout",

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Duration of the timeout",
				Key:             "duration",
				Name:            "Duration",
				Optional:        false,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeDuration,
			},
			{
				Default:         "",
				Description:     "Reason why the user was timed out",
				Key:             "reason",
				Name:            "Reason",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	args.RegisterMessageModFunc("/timeout", handleChatCommand)

	return nil
}

type actor struct{}

func (actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	reason, err := formatMessage(attrs.MustString("reason", ptrStringEmpty), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "executing reason template")
	}

	return false, errors.Wrap(
		botTwitchClient().BanUser(
			context.Background(),
			plugins.DeriveChannel(m, eventData),
			plugins.DeriveUser(m, eventData),
			attrs.MustDuration("duration", nil),
			reason,
		),
		"executing timeout",
	)
}

func (actor) IsAsync() bool { return false }
func (actor) Name() string  { return actorName }

func (actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "duration", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeDuration}),
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "reason", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.MustHaveNoUnknowFields,
		helpers.SchemaValidateTemplateField(tplValidator, "reason"),
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	if attrs.MustDuration("duration", nil) < time.Second {
		return errors.New("duration must be greater or equal one second")
	}

	return nil
}

func handleChatCommand(m *irc.Message) error {
	channel := plugins.DeriveChannel(m, nil)

	matches := timeoutChatcommandRegex.FindStringSubmatch(m.Trailing())
	if matches == nil {
		return errors.New("timeout message does not match required format")
	}

	duration, err := strconv.ParseInt(matches[2], 10, 64)
	if err != nil {
		return errors.Wrap(err, "parsing timeout duration")
	}

	if err = botTwitchClient().BanUser(context.Background(), channel, matches[1], time.Duration(duration)*time.Second, matches[3]); err != nil {
		return errors.Wrap(err, "executing timeout")
	}

	return plugins.ErrSkipSendingMessage
}

// Package shoutout contains an actor to create a Twitch native
// shoutout
package shoutout

import (
	"context"
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/go_helpers/v2/fieldcollection"
	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "shoutout"

var (
	botTwitchClient func() *twitch.Client
	formatMessage   plugins.MsgFormatter
	ptrStringEmpty  = func(v string) *string { return &v }("")

	shoutoutChatcommandRegex = regexp.MustCompile(`^/shoutout +([^\s]+)$`)
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	botTwitchClient = args.GetTwitchClient
	formatMessage = args.FormatMessage

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Perform a Twitch-native shoutout",
		Name:        "Shoutout",
		Type:        actorName,

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "User to give the shoutout to",
				Key:             "user",
				Name:            "User",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	args.RegisterMessageModFunc("/shoutout", handleChatCommand)

	return nil
}

type actor struct{}

func (actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	user, err := formatMessage(attrs.MustString("user", ptrStringEmpty), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "executing user template")
	}

	return false, errors.Wrap(
		botTwitchClient().SendShoutout(
			context.Background(),
			plugins.DeriveChannel(m, eventData),
			user,
		),
		"executing shoutout",
	)
}

func (actor) IsAsync() bool { return false }
func (actor) Name() string  { return actorName }

func (actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "user", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		helpers.SchemaValidateTemplateField(tplValidator, "user"),
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	return nil
}

func handleChatCommand(m *irc.Message) error {
	channel := plugins.DeriveChannel(m, nil)

	matches := shoutoutChatcommandRegex.FindStringSubmatch(m.Trailing())
	if matches == nil {
		return errors.New("shoutout message does not match required format")
	}

	if err := botTwitchClient().SendShoutout(context.Background(), channel, matches[1]); err != nil {
		return errors.Wrap(err, "executing shoutout")
	}

	return plugins.ErrSkipSendingMessage
}

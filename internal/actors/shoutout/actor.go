package shoutout

import (
	"regexp"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "shoutout"

var (
	botTwitchClient *twitch.Client
	formatMessage   plugins.MsgFormatter
	ptrStringEmpty  = func(v string) *string { return &v }("")

	shoutoutChatcommandRegex = regexp.MustCompile(`^/shoutout +([^\s]+)$`)
)

func Register(args plugins.RegistrationArguments) error {
	botTwitchClient = args.GetTwitchClient()
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

func (a actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	user, err := formatMessage(attrs.MustString("user", ptrStringEmpty), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "executing user template")
	}

	return false, errors.Wrap(
		botTwitchClient.SendShoutout(
			plugins.DeriveChannel(m, eventData),
			user,
		),
		"executing shoutout",
	)
}

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return actorName }

func (a actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *plugins.FieldCollection) (err error) {
	if v, err := attrs.String("user"); err != nil || v == "" {
		return errors.New("user must be non-empty string")
	}

	if err = tplValidator(attrs.MustString("user", ptrStringEmpty)); err != nil {
		return errors.Wrap(err, "validating user template")
	}

	return nil
}

func handleChatCommand(m *irc.Message) error {
	channel := plugins.DeriveChannel(m, nil)

	matches := shoutoutChatcommandRegex.FindStringSubmatch(m.Trailing())
	if matches == nil {
		return errors.New("shoutout message does not match required format")
	}

	if err := botTwitchClient.SendShoutout(channel, matches[1]); err != nil {
		return errors.Wrap(err, "executing shoutout")
	}

	return plugins.ErrSkipSendingMessage
}

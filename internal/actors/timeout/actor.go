package timeout

import (
	"regexp"
	"strconv"
	"time"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "timeout"

var (
	botTwitchClient *twitch.Client
	formatMessage   plugins.MsgFormatter
	ptrStringEmpty  = func(v string) *string { return &v }("")

	timeoutChatcommandRegex = regexp.MustCompile(`^/timeout +([^\s]+) +([0-9]+) +(.+)$`)
)

func Register(args plugins.RegistrationArguments) error {
	botTwitchClient = args.GetTwitchClient()
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

func (a actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	reason, err := formatMessage(attrs.MustString("reason", ptrStringEmpty), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "executing reason template")
	}

	return false, errors.Wrap(
		botTwitchClient.BanUser(
			plugins.DeriveChannel(m, eventData),
			plugins.DeriveUser(m, eventData),
			attrs.MustDuration("duration", nil),
			reason,
		),
		"executing timeout",
	)
}

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return actorName }

func (a actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *plugins.FieldCollection) (err error) {
	if v, err := attrs.Duration("duration"); err != nil || v < time.Second {
		return errors.New("duration must be of type duration greater or equal one second")
	}

	if v, err := attrs.String("reason"); err != nil || v == "" {
		return errors.New("reason must be non-empty string")
	}

	if err = tplValidator(attrs.MustString("reason", ptrStringEmpty)); err != nil {
		return errors.Wrap(err, "validating reason template")
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

	if err = botTwitchClient.BanUser(channel, matches[1], time.Duration(duration)*time.Second, matches[3]); err != nil {
		return errors.Wrap(err, "executing timeout")
	}

	return plugins.ErrSkipSendingMessage
}

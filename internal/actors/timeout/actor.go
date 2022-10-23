package timeout

import (
	"strconv"
	"strings"
	"time"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v2/plugins"
)

const actorName = "timeout"

var (
	formatMessage  plugins.MsgFormatter
	ptrStringEmpty = func(v string) *string { return &v }("")
)

func Register(args plugins.RegistrationArguments) error {
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
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	return nil
}

type actor struct{}

func (a actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	cmd := []string{
		"/timeout",
		plugins.DeriveUser(m, eventData),
		strconv.FormatInt(int64(attrs.MustDuration("duration", nil)/time.Second), 10),
	}

	reason, err := formatMessage(attrs.MustString("reason", ptrStringEmpty), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "executing reason template")
	}

	if reason != "" {
		cmd = append(cmd, reason)
	}

	return false, errors.Wrap(
		c.WriteMessage(&irc.Message{
			Command: "PRIVMSG",
			Params: []string{
				plugins.DeriveChannel(m, eventData),
				strings.Join(cmd, " "),
			},
		}),
		"sending timeout",
	)
}

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return actorName }

func (a actor) Validate(attrs *plugins.FieldCollection) (err error) {
	if v, err := attrs.Duration("duration"); err != nil || v < time.Second {
		return errors.New("duration must be of type duration greater or equal one second")
	}

	return nil
}

// Package commercial contains an actor to run commercials in a channel
package commercial

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/go_helpers/v2/fieldcollection"
	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const (
	actorName = "commercial"

	maxCommercialDuration = 180
)

var (
	formatMessage plugins.MsgFormatter
	permCheckFn   plugins.ChannelPermissionCheckFunc
	tcGetter      func(string) (*twitch.Client, error)

	commercialChatcommandRegex = regexp.MustCompile(`^/commercial ([0-9]+)$`)
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	formatMessage = args.FormatMessage
	permCheckFn = args.HasPermissionForChannel
	tcGetter = args.GetTwitchClientForChannel

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Start Commercial",
		Name:        "Commercial",
		Type:        actorName,

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Duration of the commercial (must not be longer than 180s and must yield an integer)",
				Key:             "duration",
				Name:            "Duration",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	args.RegisterMessageModFunc("/commercial", handleChatCommand)

	return nil
}

type actor struct{}

func (actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	ptrStringEmpty := func(v string) *string { return &v }("")

	durationStr, err := formatMessage(attrs.MustString("duration", ptrStringEmpty), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "executing duration template")
	}

	return false, startCommercial(strings.TrimLeft(plugins.DeriveChannel(m, eventData), "#"), durationStr)
}

func (actor) IsAsync() bool { return false }
func (actor) Name() string  { return actorName }

func (actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "duration", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.MustHaveNoUnknowFields,
		helpers.SchemaValidateTemplateField(tplValidator, "duration"),
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	return nil
}

func handleChatCommand(m *irc.Message) error {
	channel := strings.TrimLeft(plugins.DeriveChannel(m, nil), "#")

	matches := commercialChatcommandRegex.FindStringSubmatch(m.Trailing())
	if matches == nil {
		return errors.New("ban message does not match required format")
	}

	if err := startCommercial(channel, matches[1]); err != nil {
		return err
	}

	return plugins.ErrSkipSendingMessage
}

func startCommercial(channel, durationStr string) error {
	duration, err := strconv.ParseInt(durationStr, 10, 64)
	if err != nil {
		return errors.Wrap(err, "parsing duration to integer")
	}

	if duration > maxCommercialDuration {
		return errors.New("duration too long")
	}

	ok, err := permCheckFn(channel, twitch.ScopeChannelEditCommercial)
	if err != nil {
		return errors.Wrap(err, "checking for channel permissions")
	}

	if !ok {
		return errors.Errorf("channel %q is missing permission %s", channel, twitch.ScopeChannelEditCommercial)
	}

	tc, err := tcGetter(channel)
	if err != nil {
		return errors.Wrap(err, "getting channel twitch-client")
	}

	return errors.Wrap(
		tc.RunCommercial(context.Background(), channel, duration),
		"running commercial",
	)
}

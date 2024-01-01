// Package modchannel contains an actor to modify title / category of
// a channel
package modchannel

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "modchannel"

var (
	formatMessage plugins.MsgFormatter
	tcGetter      func(string) (*twitch.Client, error)

	ptrStringEmpty = func(s string) *string { return &s }("")
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	formatMessage = args.FormatMessage
	tcGetter = args.GetTwitchClientForChannel

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Update stream information",
		Name:        "Modify Stream",
		Type:        "modchannel",

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Channel to update",
				Key:             "channel",
				Name:            "Channel",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "",
				Description:     "Category / Game to set (use `@1234` format to pass an explicit ID)",
				Key:             "game",
				Name:            "Game",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "",
				Description:     "Stream title to set",
				Key:             "title",
				Name:            "Title",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	return nil
}

type actor struct{}

func (actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	var (
		game  = attrs.MustString("game", ptrStringEmpty)
		title = attrs.MustString("title", ptrStringEmpty)
	)

	if game == "" && title == "" {
		return false, nil
	}

	var updGame, updTitle *string

	channel, err := formatMessage(attrs.MustString("channel", nil), m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "parsing channel")
	}

	if game != "" {
		parsedGame, err := formatMessage(game, m, r, eventData)
		if err != nil {
			return false, errors.Wrap(err, "parsing game")
		}

		updGame = &parsedGame
	}

	if title != "" {
		parsedTitle, err := formatMessage(title, m, r, eventData)
		if err != nil {
			return false, errors.Wrap(err, "parsing title")
		}

		updTitle = &parsedTitle
	}

	twitchClient, err := tcGetter(strings.TrimLeft(channel, "#"))
	if err != nil {
		return false, errors.Wrap(err, "getting Twitch client")
	}

	return false, errors.Wrap(
		twitchClient.ModifyChannelInformation(context.Background(), strings.TrimLeft(channel, "#"), updGame, updTitle),
		"updating channel info",
	)
}

func (actor) IsAsync() bool { return false }
func (actor) Name() string  { return actorName }

func (actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *plugins.FieldCollection) (err error) {
	if v, err := attrs.String("channel"); err != nil || v == "" {
		return errors.New("channel must be non-empty string")
	}

	for _, field := range []string{"channel", "game", "title"} {
		if err = tplValidator(attrs.MustString(field, ptrStringEmpty)); err != nil {
			return errors.Wrapf(err, "validating %s template", field)
		}
	}

	return nil
}

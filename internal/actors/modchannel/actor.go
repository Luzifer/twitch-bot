package modchannel

import (
	"context"
	"strings"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/Luzifer/twitch-bot/twitch"
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

const actorName = "modchannel"

var (
	formatMessage plugins.MsgFormatter
	twitchClient  *twitch.Client
)

func Register(args plugins.RegistrationArguments) error {
	formatMessage = args.FormatMessage
	twitchClient = args.GetTwitchClient()

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	return nil
}

type actor struct{}

func (a actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData plugins.FieldCollection, attrs plugins.FieldCollection) (preventCooldown bool, err error) {
	var (
		ptrStringEmpty = func(v string) *string { return &v }("")
		game           = attrs.MustString("update_game", ptrStringEmpty)
		title          = attrs.MustString("update_title", ptrStringEmpty)
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

	return false, errors.Wrap(
		twitchClient.ModifyChannelInformation(context.Background(), strings.TrimLeft(channel, "#"), updGame, updTitle),
		"updating channel info",
	)
}

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return actorName }

func (a actor) Validate(attrs plugins.FieldCollection) (err error) {
	if v, err := attrs.String("channel"); err != nil || v == "" {
		return errors.New("channel must be non-empty string")
	}

	return nil
}

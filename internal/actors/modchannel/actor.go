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

type actor struct {
	Channel     string  `json:"channel" yaml:"channel"`
	UpdateGame  *string `json:"update_game" yaml:"update_game"`
	UpdateTitle *string `json:"update_title" yaml:"update_title"`
}

func (a actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData plugins.FieldCollection) (preventCooldown bool, err error) {
	if a.UpdateGame == nil && a.UpdateTitle == nil {
		return false, nil
	}

	var game, title *string

	channel, err := formatMessage(a.Channel, m, r, eventData)
	if err != nil {
		return false, errors.Wrap(err, "parsing channel")
	}

	if a.UpdateGame != nil {
		parsedGame, err := formatMessage(*a.UpdateGame, m, r, eventData)
		if err != nil {
			return false, errors.Wrap(err, "parsing game")
		}

		game = &parsedGame
	}

	if a.UpdateTitle != nil {
		parsedTitle, err := formatMessage(*a.UpdateTitle, m, r, eventData)
		if err != nil {
			return false, errors.Wrap(err, "parsing title")
		}

		title = &parsedTitle
	}

	return false, errors.Wrap(
		twitchClient.ModifyChannelInformation(context.Background(), strings.TrimLeft(channel, "#"), game, title),
		"updating channel info",
	)
}

func (a actor) IsAsync() bool { return false }
func (a actor) Name() string  { return actorName }

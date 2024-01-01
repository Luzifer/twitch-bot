// Package announce contains a chat essage handler to create
// announcements from the bot
package announce

import (
	"context"
	"regexp"

	"github.com/pkg/errors"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

var (
	botTwitchClient *twitch.Client

	announceChatcommandRegex = regexp.MustCompile(`^/announce(|blue|green|orange|purple) +(.+)$`)
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	botTwitchClient = args.GetTwitchClient()

	args.RegisterMessageModFunc("/announce", handleChatCommand)

	return nil
}

func handleChatCommand(m *irc.Message) error {
	channel := plugins.DeriveChannel(m, nil)

	matches := announceChatcommandRegex.FindStringSubmatch(m.Trailing())
	if matches == nil {
		return errors.New("announce message does not match required format")
	}

	if err := botTwitchClient.SendChatAnnouncement(context.Background(), channel, matches[1], matches[2]); err != nil {
		return errors.Wrap(err, "sending announcement")
	}

	return plugins.ErrSkipSendingMessage
}

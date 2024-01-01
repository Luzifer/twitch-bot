package raffle

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

func rawMessageHandler(m *irc.Message) error {
	if m.Command != "PRIVMSG" {
		// We only care for messages containing the raffle keyword
		return nil
	}

	channel := plugins.DeriveChannel(m, nil)
	if channel == "" {
		// The frick? Messages should have a channel but whatever
		return nil
	}

	user := plugins.DeriveUser(m, nil)
	if user == "" {
		// The frick? Messages should have a user but whatever
		return nil
	}

	go func() {
		if err := dbc.RegisterSpeakUp(channel, user, m.Trailing()); err != nil {
			logrus.WithFields(logrus.Fields{
				"channel": channel,
				"user":    user,
			}).WithError(err).Error("registering speak-up")
		}
	}()

	return handleRaffleEntry(m, channel, user)
}

//nolint:gocyclo // Dividing would need to carry over everyhing and make it more complex
func handleRaffleEntry(m *irc.Message, channel, user string) error {
	flds := strings.Fields(m.Trailing())
	if len(flds) == 0 {
		// A message also should have: a message!
		return nil
	}

	var (
		badges     = twitch.ParseBadgeLevels(m)
		doesFollow bool
		keyword    = flds[0]
	)

	r, err := dbc.GetByChannelAndKeyword(channel, keyword)
	if err != nil {
		if errors.Is(err, errRaffleNotFound) {
			// We don't need to care, that was no raffle input
			return nil
		}
		return errors.Wrap(err, "fetching raffle")
	}

	raffleChan, err := tcGetter(r.Channel)
	if err != nil {
		return errors.Wrap(err, "getting twitch client for raffle")
	}

	since, err := raffleChan.GetFollowDate(context.Background(), user, strings.TrimLeft(channel, "#"))
	switch {
	case err == nil:
		doesFollow = since.Before(time.Now().Add(-r.MinFollowAge))

	case errors.Is(err, twitch.ErrUserDoesNotFollow):
		doesFollow = false

	default:
		return errors.Wrap(err, "checking follow for user")
	}

	re := raffleEntry{
		RaffleID:        r.ID,
		UserID:          m.Tags["user-id"],
		UserLogin:       user,
		UserDisplayName: m.Tags["display-name"],
		EnteredAt:       time.Now().UTC(),
	}

	if re.UserDisplayName == "" {
		re.UserDisplayName = re.UserLogin
	}

	raffleEventFields := plugins.FieldCollectionFromData(map[string]any{
		"user_id": m.Tags["user-id"],
		"user":    user,
	})

	switch {
	case r.AllowVIP && badges.Has(twitch.BadgeVIP):
		re.EnteredAs = twitch.BadgeVIP
		re.Multiplier = r.MultiVIP

	case r.AllowSubscriber && badges.Has(twitch.BadgeSubscriber):
		re.EnteredAs = twitch.BadgeSubscriber
		re.Multiplier = r.MultiSubscriber

	case r.AllowFollower && doesFollow:
		re.EnteredAs = "follower"
		re.Multiplier = r.MultiFollower

	case r.AllowEveryone:
		re.EnteredAs = "everyone"
		re.Multiplier = 1

	default:
		// Well. No luck, no entry.
		return errors.Wrap(
			r.SendEvent(raffleMessageEventEntryFailed, raffleEventFields),
			"sending entry-failed chat message",
		)
	}

	// We have everything we need to create an entry
	if err = dbc.Enter(re); err != nil {
		logrus.WithFields(logrus.Fields{
			"raffle":  r.ID,
			"user_id": re.UserID,
			"user":    re.UserLogin,
		}).WithError(err).Error("creating raffle entry")
		return errors.Wrap(
			r.SendEvent(raffleMessageEventEntryFailed, raffleEventFields),
			"sending entry-failed chat message",
		)
	}

	return errors.Wrap(
		r.SendEvent(raffleMessageEventEntry, raffleEventFields),
		"sending entry chat message",
	)
}

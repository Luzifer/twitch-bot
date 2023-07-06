package raffle

import (
	"strings"
	"time"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

//nolint:funlen,gocyclo // Dividing would need to carry over everyhing and make it more complex
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

	since, err := raffleChan.GetFollowDate(plugins.DeriveUser(m, nil), strings.TrimLeft(channel, "#"))
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
		UserID:          string(m.Tags["user-id"]),
		UserLogin:       plugins.DeriveUser(m, nil),
		UserDisplayName: string(m.Tags["display-name"]),
		EnteredAt:       time.Now().UTC(),
	}

	if re.UserDisplayName == "" {
		re.UserDisplayName = re.UserLogin
	}

	raffleEventFields := plugins.FieldCollectionFromData(map[string]any{
		"user_id": string(m.Tags["user-id"]),
		"user":    plugins.DeriveUser(m, nil),
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

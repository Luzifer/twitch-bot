package raffle

import (
	"time"

	"github.com/Luzifer/twitch-bot/v3/plugins"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/irc.v4"
)

type (
	enterRaffleActor struct{}
)

var ptrStrEmpty = ptrStr("")

func ptrStr(v string) *string { return &v }

func (a enterRaffleActor) Execute(_ *irc.Client, m *irc.Message, _ *plugins.Rule, evtData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	if m != nil || evtData.MustString("reward_id", ptrStrEmpty) == "" {
		return false, errors.New("enter-raffle actor is only supposed to act on channelpoint redeems")
	}

	r, err := dbc.GetByChannelAndKeyword(
		evtData.MustString("channel", ptrStrEmpty),
		attrs.MustString("keyword", ptrStrEmpty),
	)
	if err != nil {
		if errors.Is(err, errRaffleNotFound) {
			// We don't need to care, that was no raffle input
			return false, errors.Errorf("specified keyword %q does not belong to active raffle", attrs.MustString("keyword", ptrStrEmpty))
		}
		return false, errors.Wrap(err, "fetching raffle")
	}

	re := raffleEntry{
		EnteredAs:       "reward",
		RaffleID:        r.ID,
		UserID:          evtData.MustString("user_id", ptrStrEmpty),
		UserLogin:       evtData.MustString("user", ptrStrEmpty),
		UserDisplayName: evtData.MustString("user", ptrStrEmpty),
		EnteredAt:       time.Now().UTC(),
	}

	raffleEventFields := plugins.FieldCollectionFromData(map[string]any{
		"user_id": re.UserID,
		"user":    re.UserLogin,
	})

	// We have everything we need to create an entry
	if err = dbc.Enter(re); err != nil {
		logrus.WithFields(logrus.Fields{
			"raffle":  r.ID,
			"user_id": re.UserID,
			"user":    re.UserLogin,
		}).WithError(err).Error("creating raffle entry")
		return false, errors.Wrap(
			r.SendEvent(raffleMessageEventEntryFailed, raffleEventFields),
			"sending entry-failed chat message",
		)
	}

	return false, errors.Wrap(
		r.SendEvent(raffleMessageEventEntry, raffleEventFields),
		"sending entry chat message",
	)
}

func (a enterRaffleActor) IsAsync() bool { return false }
func (a enterRaffleActor) Name() string  { return "enter-raffle" }

func (a enterRaffleActor) Validate(_ plugins.TemplateValidatorFunc, attrs *plugins.FieldCollection) (err error) {
	keyword, err := attrs.String("keyword")
	if err != nil || keyword == "" {
		return errors.New("keyword must be non-empty string")
	}

	return nil
}

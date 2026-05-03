package raffle

import (
	"errors"
	"fmt"
	"time"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"github.com/sirupsen/logrus"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

type (
	enterRaffleActor struct{}
)

var ptrStrEmpty = ptrStr("")

func ptrStr(v string) *string { return &v }

func (enterRaffleActor) Execute(_ *irc.Client, m *irc.Message, _ *plugins.Rule, evtData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
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
			return false, fmt.Errorf("specified keyword %q does not belong to active raffle", attrs.MustString("keyword", ptrStrEmpty))
		}
		return false, fmt.Errorf("fetching raffle: %w", err)
	}

	re := raffleEntry{
		EnteredAs:       "reward",
		RaffleID:        r.ID,
		UserID:          evtData.MustString("user_id", ptrStrEmpty),
		UserLogin:       evtData.MustString("user", ptrStrEmpty),
		UserDisplayName: evtData.MustString("user", ptrStrEmpty),
		EnteredAt:       time.Now().UTC(),
	}

	raffleEventFields := fieldcollection.FieldCollectionFromData(map[string]any{
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

		if err = r.SendEvent(raffleMessageEventEntryFailed, raffleEventFields); err != nil {
			return false, fmt.Errorf("sending entry-failed chat message: %w", err)
		}

		// Entry failed but we notified about that, so there is no need
		// for us to error to the outside: This can happen on duplicate
		// entries or other reasons.
		return false, nil
	}

	if err = r.SendEvent(raffleMessageEventEntry, raffleEventFields); err != nil {
		return false, fmt.Errorf("sending entry chat message: %w", err)
	}

	return false, nil
}

func (enterRaffleActor) IsAsync() bool { return false }
func (enterRaffleActor) Name() string  { return "enter-raffle" }

func (enterRaffleActor) Validate(_ plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "keyword", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.MustHaveNoUnknowFields,
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	return nil
}

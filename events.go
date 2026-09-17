package main

import (
	"sync"

	"github.com/Luzifer/go_helpers/fieldcollection"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v3/pkg/event"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

var (
	registeredEventHandlers     []plugins.EventHandlerFunc
	registeredEventHandlersLock sync.Mutex
)

var (
	eventTypeCustom       = new("custom")
	eventTypeKoFiDonation = new("kofi_donation")

	knownEvents = []*string{
		event.AdBreakBegin{}.Event(),
		event.Announcement{}.Event(),
		event.Ban{}.Event(),
		event.Bits{}.Event(),
		eventTypeCustom,
		event.ChannelPointRedeem{}.Event(),
		event.ClearChat{}.Event(),
		event.ClearMessage{}.Event(),
		event.Follow{}.Event(),
		event.GiftPaidUpgrade{}.Event(),
		event.HypetrainBegin{}.Event(),
		event.HypetrainEnd{}.Event(),
		event.HypetrainProgress{}.Event(),
		event.Join{}.Event(),
		eventTypeKoFiDonation,
		event.ModeratorAdd{}.Event(),
		event.ModeratorRemove{}.Event(),
		event.OutboundRaid{}.Event(),
		event.Part{}.Event(),
		event.Permit{}.Event(),
		event.PollBegin{}.Event(),
		event.PollEnd{}.Event(),
		event.PollProgress{}.Event(),
		event.PrimePaidUpgrade{}.Event(),
		event.Raid{}.Event(),
		event.Resub{}.Event(),
		event.ShoutoutCreated{}.Event(),
		event.ShoutoutReceived{}.Event(),
		event.Sub{}.Event(),
		event.SubGift{}.Event(),
		event.SubMysteryGift{}.Event(),
		event.SuspiciousUserMessage{}.Event(),
		event.SuspiciousUserUpdate{}.Event(),
		event.Timeout{}.Event(),
		event.VIPAdd{}.Event(),
		event.VIPRemove{}.Event(),
		event.WatchStreak{}.Event(),
		event.Whisper{}.Event(),

		event.CategoryUpdate{}.Event(),
		event.StreamOffline{}.Event(),
		event.StreamOnline{}.Event(),
		event.TitleUpdate{}.Event(),
	}
)

func notifyEventHandlers(evt string, eventData *fieldcollection.FieldCollection) {
	registeredEventHandlersLock.Lock()
	defer registeredEventHandlersLock.Unlock()

	for _, fn := range registeredEventHandlers {
		if err := fn(evt, eventData); err != nil {
			log.WithError(err).Error("EventHandler caused error")
		}
	}
}

func registerEventHandlers(eh plugins.EventHandlerFunc) error {
	registeredEventHandlersLock.Lock()
	defer registeredEventHandlersLock.Unlock()

	registeredEventHandlers = append(registeredEventHandlers, eh)
	return nil
}

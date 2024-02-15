package main

import (
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

func ptrStr(s string) *string { return &s }

var (
	registeredEventHandlers     []plugins.EventHandlerFunc
	registeredEventHandlersLock sync.Mutex
)

var (
	eventTypeAdBreakBegin       = ptrStr("adbreak_begin")
	eventTypeAnnouncement       = ptrStr("announcement")
	eventTypeBan                = ptrStr("ban")
	eventTypeBits               = ptrStr("bits")
	eventTypeCustom             = ptrStr("custom")
	eventTypeChannelPointRedeem = ptrStr("channelpoint_redeem")
	eventTypeClearChat          = ptrStr("clearchat")
	eventTypeDelete             = ptrStr("delete")
	eventTypeFollow             = ptrStr("follow")
	eventTypeGiftPaidUpgrade    = ptrStr("giftpaidupgrade")
	eventTypeJoin               = ptrStr("join")
	eventKoFiDonation           = ptrStr("kofi_donation")
	eventTypeOutboundRaid       = ptrStr("outbound_raid")
	eventTypePart               = ptrStr("part")
	eventTypePermit             = ptrStr("permit")
	eventTypePollBegin          = ptrStr("poll_begin")
	eventTypePollEnd            = ptrStr("poll_end")
	eventTypePollProgress       = ptrStr("poll_progress")
	eventTypeRaid               = ptrStr("raid")
	eventTypeResub              = ptrStr("resub")
	eventTypeShoutoutCreated    = ptrStr("shoutout_created")
	eventTypeShoutoutReceived   = ptrStr("shoutout_received")
	eventTypeSubgift            = ptrStr("subgift")
	eventTypeSubmysterygift     = ptrStr("submysterygift")
	eventTypeSub                = ptrStr("sub")
	eventTypeTimeout            = ptrStr("timeout")
	eventTypeWatchStreak        = ptrStr("watch_streak")
	eventTypeWhisper            = ptrStr("whisper")

	eventTypeTwitchCategoryUpdate = ptrStr("category_update")
	eventTypeTwitchStreamOffline  = ptrStr("stream_offline")
	eventTypeTwitchStreamOnline   = ptrStr("stream_online")
	eventTypeTwitchTitleUpdate    = ptrStr("title_update")

	knownEvents = []*string{
		eventTypeAdBreakBegin,
		eventTypeAnnouncement,
		eventTypeBan,
		eventTypeBits,
		eventTypeCustom,
		eventTypeChannelPointRedeem,
		eventTypeClearChat,
		eventTypeDelete,
		eventTypeFollow,
		eventTypeGiftPaidUpgrade,
		eventTypeJoin,
		eventKoFiDonation,
		eventTypeOutboundRaid,
		eventTypePart,
		eventTypePermit,
		eventTypePollBegin,
		eventTypePollEnd,
		eventTypePollProgress,
		eventTypeRaid,
		eventTypeResub,
		eventTypeShoutoutCreated,
		eventTypeShoutoutReceived,
		eventTypeSub,
		eventTypeSubgift,
		eventTypeSubmysterygift,
		eventTypeTimeout,
		eventTypeWatchStreak,
		eventTypeWhisper,

		eventTypeTwitchCategoryUpdate,
		eventTypeTwitchStreamOffline,
		eventTypeTwitchStreamOnline,
		eventTypeTwitchTitleUpdate,
	}
)

func notifyEventHandlers(event string, eventData *plugins.FieldCollection) {
	registeredEventHandlersLock.Lock()
	defer registeredEventHandlersLock.Unlock()

	for _, fn := range registeredEventHandlers {
		if err := fn(event, eventData); err != nil {
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

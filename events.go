package main

import (
	"sync"

	"github.com/Luzifer/go_helpers/fieldcollection"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

var (
	registeredEventHandlers     []plugins.EventHandlerFunc
	registeredEventHandlersLock sync.Mutex
)

var (
	eventTypeAdBreakBegin       = new("adbreak_begin")
	eventTypeAnnouncement       = new("announcement")
	eventTypeBan                = new("ban")
	eventTypeBits               = new("bits")
	eventTypeCustom             = new("custom")
	eventTypeChannelPointRedeem = new("channelpoint_redeem")
	eventTypeClearChat          = new("clearchat")
	eventTypeDelete             = new("delete")
	eventTypeFollow             = new("follow")
	eventTypeGiftPaidUpgrade    = new("giftpaidupgrade")
	eventTypeHypetrainBegin     = new("hypetrain_begin")
	eventTypeHypetrainEnd       = new("hypetrain_end")
	eventTypeHypetrainProgress  = new("hypetrain_progress")
	eventTypeJoin               = new("join")
	eventKoFiDonation           = new("kofi_donation")
	eventTypeOutboundRaid       = new("outbound_raid")
	eventTypePart               = new("part")
	eventTypePermit             = new("permit")
	eventTypePollBegin          = new("poll_begin")
	eventTypePollEnd            = new("poll_end")
	eventTypePollProgress       = new("poll_progress")
	eventTypeRaid               = new("raid")
	eventTypeResub              = new("resub")
	eventTypeShoutoutCreated    = new("shoutout_created")
	eventTypeShoutoutReceived   = new("shoutout_received")
	eventTypeSubgift            = new("subgift")
	eventTypeSubmysterygift     = new("submysterygift")
	eventTypeSub                = new("sub")
	eventTypeSusUserMessage     = new("sus_user_message")
	eventTypeSusUserUpdate      = new("sus_user_update")
	eventTypeTimeout            = new("timeout")
	eventTypeWatchStreak        = new("watch_streak")
	eventTypeWhisper            = new("whisper")

	eventTypeTwitchCategoryUpdate = new("category_update")
	eventTypeTwitchStreamOffline  = new("stream_offline")
	eventTypeTwitchStreamOnline   = new("stream_online")
	eventTypeTwitchTitleUpdate    = new("title_update")

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
		eventTypeHypetrainBegin,
		eventTypeHypetrainEnd,
		eventTypeHypetrainProgress,
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
		eventTypeSusUserMessage,
		eventTypeSusUserUpdate,
		eventTypeTimeout,
		eventTypeWatchStreak,
		eventTypeWhisper,

		eventTypeTwitchCategoryUpdate,
		eventTypeTwitchStreamOffline,
		eventTypeTwitchStreamOnline,
		eventTypeTwitchTitleUpdate,
	}
)

func notifyEventHandlers(event string, eventData *fieldcollection.FieldCollection) {
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

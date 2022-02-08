package main

import (
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/plugins"
)

func ptrStr(s string) *string { return &s }

var (
	registeredEventHandlers     []plugins.EventHandlerFunc
	registeredEventHandlersLock sync.Mutex
)

var (
	eventTypeBan                = ptrStr("ban")
	eventTypeBits               = ptrStr("bits")
	eventTypeChannelPointRedeem = ptrStr("channelpoint_redeem")
	eventTypeClearChat          = ptrStr("clearchat")
	eventTypeFollow             = ptrStr("follow")
	eventTypeGiftPaidUpgrade    = ptrStr("giftpaidupgrade")
	eventTypeHost               = ptrStr("host")
	eventTypeJoin               = ptrStr("join")
	eventTypePart               = ptrStr("part")
	eventTypePermit             = ptrStr("permit")
	eventTypeRaid               = ptrStr("raid")
	eventTypeResub              = ptrStr("resub")
	eventTypeSubgift            = ptrStr("subgift")
	eventTypeSubmysterygift     = ptrStr("submysterygift")
	eventTypeSub                = ptrStr("sub")
	eventTypeTimeout            = ptrStr("timeout")
	eventTypeWhisper            = ptrStr("whisper")

	eventTypeTwitchCategoryUpdate = ptrStr("category_update")
	eventTypeTwitchStreamOffline  = ptrStr("stream_offline")
	eventTypeTwitchStreamOnline   = ptrStr("stream_online")
	eventTypeTwitchTitleUpdate    = ptrStr("title_update")

	knownEvents = []*string{
		eventTypeBan,
		eventTypeBits,
		eventTypeChannelPointRedeem,
		eventTypeClearChat,
		eventTypeFollow,
		eventTypeGiftPaidUpgrade,
		eventTypeHost,
		eventTypeJoin,
		eventTypePart,
		eventTypePermit,
		eventTypeRaid,
		eventTypeResub,
		eventTypeSub,
		eventTypeSubgift,
		eventTypeSubmysterygift,
		eventTypeTimeout,
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

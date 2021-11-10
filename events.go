package main

func ptrStr(s string) *string { return &s }

var (
	eventTypeBan            = ptrStr("ban")
	eventTypeBits           = ptrStr("bits")
	eventTypeClearChat      = ptrStr("clearchat")
	eventTypeJoin           = ptrStr("join")
	eventTypeHost           = ptrStr("host")
	eventTypePart           = ptrStr("part")
	eventTypePermit         = ptrStr("permit")
	eventTypeRaid           = ptrStr("raid")
	eventTypeResub          = ptrStr("resub")
	eventTypeSub            = ptrStr("sub")
	eventTypeSubgift        = ptrStr("subgift")
	eventTypeSubmysterygift = ptrStr("submysterygift")
	eventTypeTimeout        = ptrStr("timeout")
	eventTypeWhisper        = ptrStr("whisper")

	eventTypeTwitchCategoryUpdate = ptrStr("category_update")
	eventTypeTwitchStreamOffline  = ptrStr("stream_offline")
	eventTypeTwitchStreamOnline   = ptrStr("stream_online")
	eventTypeTwitchTitleUpdate    = ptrStr("title_update")

	knownEvents = []*string{
		eventTypeBan,
		eventTypeBits,
		eventTypeClearChat,
		eventTypeJoin,
		eventTypeHost,
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

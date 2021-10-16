package main

func ptrStr(s string) *string { return &s }

var (
	eventTypeJoin           = ptrStr("join")
	eventTypeHost           = ptrStr("host")
	eventTypePart           = ptrStr("part")
	eventTypePermit         = ptrStr("permit")
	eventTypeRaid           = ptrStr("raid")
	eventTypeResub          = ptrStr("resub")
	eventTypeSub            = ptrStr("sub")
	eventTypeSubgift        = ptrStr("subgift")
	eventTypeSubmysterygift = ptrStr("submysterygift")
	eventTypeWhisper        = ptrStr("whisper")

	eventTypeTwitchCategoryUpdate = ptrStr("category_update")
	eventTypeTwitchStreamOffline  = ptrStr("stream_offline")
	eventTypeTwitchStreamOnline   = ptrStr("stream_online")
	eventTypeTwitchTitleUpdate    = ptrStr("title_update")

	knownEvents = []*string{
		eventTypeJoin,
		eventTypeHost,
		eventTypePart,
		eventTypePermit,
		eventTypeRaid,
		eventTypeResub,
		eventTypeSub,
		eventTypeSubgift,
		eventTypeSubmysterygift,
		eventTypeWhisper,

		eventTypeTwitchCategoryUpdate,
		eventTypeTwitchStreamOffline,
		eventTypeTwitchStreamOnline,
		eventTypeTwitchTitleUpdate,
	}
)

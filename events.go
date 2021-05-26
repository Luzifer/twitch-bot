package main

func ptrStr(s string) *string { return &s }

var (
	eventTypeJoin    = ptrStr("join")
	eventTypeHost    = ptrStr("host")
	eventTypePart    = ptrStr("part")
	eventTypePermit  = ptrStr("permit")
	eventTypeRaid    = ptrStr("raid")
	eventTypeResub   = ptrStr("resub")
	eventTypeSub     = ptrStr("sub")
	eventTypeSubgift = ptrStr("subgift")
	eventTypeWhisper = ptrStr("whisper")
)

package main

func ptrStr(s string) *string { return &s }

var (
	eventTypeHost   = ptrStr("host")
	eventTypePermit = ptrStr("permit")
	eventTypeRaid   = ptrStr("raid")
	eventTypeResub  = ptrStr("resub")
)

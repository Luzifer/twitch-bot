package main

var (
	ptrIntZero     = func(v int64) *int64 { return &v }(0)
	ptrStringEmpty = func(v string) *string { return &v }("")
)

package main

var (
	ptrBoolFalse   = func(v bool) *bool { return &v }(false)
	ptrIntZero     = func(v int64) *int64 { return &v }(0)
	ptrStringEmpty = func(v string) *string { return &v }("")
)

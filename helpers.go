package main

var (
	ptrBoolFalse   = func(v bool) *bool { return &v }(false)
	ptrStringEmpty = func(v string) *string { return &v }("")
)

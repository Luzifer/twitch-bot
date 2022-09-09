package main

var ptrBoolFalse = func(v bool) *bool { return &v }(false)

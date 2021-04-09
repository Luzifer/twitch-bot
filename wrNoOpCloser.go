package main

import "io"

type writeNoOpCloser struct{ io.Writer }

func (writeNoOpCloser) Close() error { return nil }

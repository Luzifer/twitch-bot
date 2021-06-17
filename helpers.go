package main

import "time"

func fixDurationValue(d time.Duration) time.Duration {
	if d >= time.Second {
		return d
	}

	return d * time.Second
}

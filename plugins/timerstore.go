package plugins

import "time"

type TimerType uint8

const (
	TimerTypePermit TimerType = iota
	TimerTypeCooldown
)

type (
	TimerEntry struct {
		Kind TimerType `json:"kind"`
		Time time.Time `json:"time"`
	}

	TimerStore interface {
		AddCooldown(tt TimerType, limiter, ruleID string, expiry time.Time)
		InCooldown(tt TimerType, limiter, ruleID string) bool
		AddPermit(channel, username string)
		HasPermit(channel, username string) bool
	}
)

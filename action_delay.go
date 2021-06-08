package main

import (
	"math/rand"
	"time"

	"github.com/go-irc/irc"
)

func init() {
	registerAction(func() Actor { return &ActorDelay{} })
}

type ActorDelay struct {
	Delay       time.Duration `json:"delay" yaml:"delay"`
	DelayJitter time.Duration `json:"delay_jitter" yaml:"delay_jitter"`
}

func (a ActorDelay) Execute(c *irc.Client, m *irc.Message, r *Rule) error {
	if a.Delay == 0 && a.DelayJitter == 0 {
		return nil
	}

	totalDelay := a.Delay
	if a.DelayJitter > 0 {
		totalDelay += time.Duration(rand.Int63n(int64(a.DelayJitter))) // #nosec: G404 // It's just time, no need for crypto/rand
	}

	time.Sleep(totalDelay)
	return nil
}

func (a ActorDelay) IsAsync() bool { return false }
func (a ActorDelay) Name() string  { return "delay" }

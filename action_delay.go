package main

import (
	"math/rand"
	"time"

	"github.com/go-irc/irc"
)

func init() {
	registerAction(func(c *irc.Client, m *irc.Message, ruleDef *rule, r *ruleAction) error {
		if r.Delay == 0 && r.DelayJitter == 0 {
			return nil
		}

		totalDelay := r.Delay
		if r.DelayJitter > 0 {
			totalDelay += time.Duration(rand.Int63n(int64(r.DelayJitter)))
		}

		time.Sleep(totalDelay)
		return nil
	})
}

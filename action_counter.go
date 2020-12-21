package main

import (
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func init() {
	registerAction(func(c *irc.Client, m *irc.Message, r *ruleAction) error {
		if r.Counter == nil {
			return nil
		}

		var counterStep int64 = 1
		if r.CounterStep != nil {
			counterStep = *r.CounterStep
		}

		return errors.Wrap(
			store.UpdateCounter(*r.Counter, counterStep, false),
			"update counter",
		)
	})
}

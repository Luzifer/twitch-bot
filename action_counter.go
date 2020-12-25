package main

import (
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func init() {
	registerAction(func(c *irc.Client, m *irc.Message, ruleDef *rule, r *ruleAction) error {
		if r.Counter == nil {
			return nil
		}

		counterName, err := formatMessage(*r.Counter, m, ruleDef, nil)
		if err != nil {
			return errors.Wrap(err, "preparing response")
		}

		var counterStep int64 = 1
		if r.CounterStep != nil {
			counterStep = *r.CounterStep
		}

		return errors.Wrap(
			store.UpdateCounter(counterName, counterStep, false),
			"update counter",
		)
	})
}

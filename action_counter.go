package main

import (
	"strconv"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func init() {
	registerAction(func(c *irc.Client, m *irc.Message, ruleDef *Rule, r *RuleAction) error {
		if r.Counter == nil {
			return nil
		}

		counterName, err := formatMessage(*r.Counter, m, ruleDef, nil)
		if err != nil {
			return errors.Wrap(err, "preparing response")
		}

		if r.CounterSet != nil {
			parseValue, err := formatMessage(*r.CounterSet, m, ruleDef, nil)
			if err != nil {
				return errors.Wrap(err, "execute counter value template")
			}

			counterValue, err := strconv.ParseInt(parseValue, 10, 64)
			if err != nil {
				return errors.Wrap(err, "parse counter value")
			}

			return errors.Wrap(
				store.UpdateCounter(counterName, counterValue, true),
				"set counter",
			)
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

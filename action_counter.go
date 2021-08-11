package main

import (
	"strconv"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func init() {
	registerAction(func() Actor { return &ActorCounter{} })
}

type ActorCounter struct {
	CounterSet  *string `json:"counter_set" yaml:"counter_set"`
	CounterStep *int64  `json:"counter_step" yaml:"counter_step"`
	Counter     *string `json:"counter" yaml:"counter"`
}

func (a ActorCounter) Execute(c *irc.Client, m *irc.Message, r *Rule) (preventCooldown bool, err error) {
	if a.Counter == nil {
		return false, nil
	}

	counterName, err := formatMessage(*a.Counter, m, r, nil)
	if err != nil {
		return false, errors.Wrap(err, "preparing response")
	}

	if a.CounterSet != nil {
		parseValue, err := formatMessage(*a.CounterSet, m, r, nil)
		if err != nil {
			return false, errors.Wrap(err, "execute counter value template")
		}

		counterValue, err := strconv.ParseInt(parseValue, 10, 64) //nolint:gomnd // Those numbers are static enough
		if err != nil {
			return false, errors.Wrap(err, "parse counter value")
		}

		return false, errors.Wrap(
			store.UpdateCounter(counterName, counterValue, true),
			"set counter",
		)
	}

	var counterStep int64 = 1
	if a.CounterStep != nil {
		counterStep = *a.CounterStep
	}

	return false, errors.Wrap(
		store.UpdateCounter(counterName, counterStep, false),
		"update counter",
	)
}

func (a ActorCounter) IsAsync() bool { return false }
func (a ActorCounter) Name() string  { return "counter" }

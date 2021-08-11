package main

import (
	"sync"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type (
	Actor interface {
		// Execute will be called after the config was read into the Actor
		Execute(*irc.Client, *irc.Message, *Rule) (preventCooldown bool, err error)
		// IsAsync may return true if the Execute function is to be executed
		// in a Go routine as of long runtime. Normally it should return false
		// except in very specific cases
		IsAsync() bool
		// Name must return an unique name for the actor in order to identify
		// it in the logs for debugging purposes
		Name() string
	}
	ActorCreationFunc func() Actor
)

var (
	availableActions     []ActorCreationFunc
	availableActionsLock = new(sync.RWMutex)
)

func registerAction(af ActorCreationFunc) {
	availableActionsLock.Lock()
	defer availableActionsLock.Unlock()

	availableActions = append(availableActions, af)
}

func triggerActions(c *irc.Client, m *irc.Message, rule *Rule, ra *RuleAction) (preventCooldown bool, err error) {
	availableActionsLock.RLock()
	defer availableActionsLock.RUnlock()

	for _, acf := range availableActions {
		var (
			a      = acf()
			logger = log.WithField("actor", a.Name())
		)

		if err := ra.Unmarshal(a); err != nil {
			logger.WithError(err).Trace("Unable to unmarshal config")
			continue
		}

		if a.IsAsync() {
			go func() {
				if _, err := a.Execute(c, m, rule); err != nil {
					logger.WithError(err).Error("Error in async actor")
				}
			}()
			continue
		}

		apc, err := a.Execute(c, m, rule)
		preventCooldown = preventCooldown || apc
		if err != nil {
			return preventCooldown, errors.Wrap(err, "execute action")
		}
	}

	return preventCooldown, nil
}

func handleMessage(c *irc.Client, m *irc.Message, event *string) {
	for _, r := range config.GetMatchingRules(m, event) {
		var preventCooldown bool

		for _, a := range r.Actions {
			apc, err := triggerActions(c, m, r, a)
			if err != nil {
				log.WithError(err).Error("Unable to trigger action")
			}
			preventCooldown = preventCooldown || apc
		}

		// Lock command
		if !preventCooldown {
			r.setCooldown(m)
		}
	}
}

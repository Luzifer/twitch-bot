package main

import (
	"sync"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	availableActions     []actionFunc
	availableActionsLock = new(sync.RWMutex)
)

type actionFunc func(*irc.Client, *irc.Message, *rule, *ruleAction) error

func registerAction(af actionFunc) {
	availableActionsLock.Lock()
	defer availableActionsLock.Unlock()

	availableActions = append(availableActions, af)
}

func triggerActions(c *irc.Client, m *irc.Message, rule *rule, ra *ruleAction) error {
	availableActionsLock.RLock()
	defer availableActionsLock.RUnlock()

	for _, af := range availableActions {
		if err := af(c, m, rule, ra); err != nil {
			return errors.Wrap(err, "execute action")
		}
	}

	return nil
}

func handleMessage(c *irc.Client, m *irc.Message, event *string) {
	for _, r := range config.GetMatchingRules(m, event) {
		for _, a := range r.Actions {
			if err := triggerActions(c, m, r, a); err != nil {
				log.WithError(err).Error("Unable to trigger action")
			}
		}

		// Lock command
		r.SetCooldown()
	}
}

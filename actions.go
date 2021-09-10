package main

import (
	"sync"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	availableActions     []plugins.ActorCreationFunc
	availableActionsLock = new(sync.RWMutex)
)

// Compile-time assertion
var _ plugins.ActorRegistrationFunc = registerAction

func registerAction(af plugins.ActorCreationFunc) {
	availableActionsLock.Lock()
	defer availableActionsLock.Unlock()

	availableActions = append(availableActions, af)
}

func triggerActions(c *irc.Client, m *irc.Message, rule *plugins.Rule, ra *plugins.RuleAction, eventData plugins.FieldCollection) (preventCooldown bool, err error) {
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
				if _, err := a.Execute(c, m, rule, eventData); err != nil {
					logger.WithError(err).Error("Error in async actor")
				}
			}()
			continue
		}

		apc, err := a.Execute(c, m, rule, eventData)
		preventCooldown = preventCooldown || apc
		if err != nil {
			return preventCooldown, errors.Wrap(err, "execute action")
		}
	}

	return preventCooldown, nil
}

func handleMessage(c *irc.Client, m *irc.Message, event *string, eventData plugins.FieldCollection) {
	for _, r := range config.GetMatchingRules(m, event, eventData) {
		var preventCooldown bool

		for _, a := range r.Actions {
			apc, err := triggerActions(c, m, r, a, eventData)
			if err != nil {
				log.WithError(err).Error("Unable to trigger action")
				return // Break execution when one action fails
			}
			preventCooldown = preventCooldown || apc
		}

		// Lock command
		if !preventCooldown {
			r.SetCooldown(timerStore, m, eventData)
		}
	}
}

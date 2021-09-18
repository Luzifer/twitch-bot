package main

import (
	"sync"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	availableActions     = map[string]plugins.ActorCreationFunc{}
	availableActionsLock = new(sync.RWMutex)
)

// Compile-time assertion
var _ plugins.ActorRegistrationFunc = registerAction

func getActorByName(name string) (plugins.Actor, error) {
	availableActionsLock.RLock()
	defer availableActionsLock.RUnlock()

	acf, ok := availableActions[name]
	if !ok {
		return nil, errors.Errorf("undefined actor %q called", name)
	}

	return acf(), nil
}

func registerAction(name string, acf plugins.ActorCreationFunc) {
	availableActionsLock.Lock()
	defer availableActionsLock.Unlock()

	if _, ok := availableActions[name]; ok {
		log.WithField("name", name).Fatal("Duplicate registration of actor")
	}

	availableActions[name] = acf
}

func triggerAction(c *irc.Client, m *irc.Message, rule *plugins.Rule, ra *plugins.RuleAction, eventData plugins.FieldCollection) (preventCooldown bool, err error) {
	availableActionsLock.RLock()
	defer availableActionsLock.RUnlock()

	a, err := getActorByName(ra.Type)
	if err != nil {
		return false, errors.Wrap(err, "getting actor")
	}

	logger := log.WithField("actor", a.Name())

	if a.IsAsync() {
		go func() {
			if _, err := a.Execute(c, m, rule, eventData, ra.Attributes); err != nil {
				logger.WithError(err).Error("Error in async actor")
			}
		}()
		return preventCooldown, nil
	}

	apc, err := a.Execute(c, m, rule, eventData, ra.Attributes)
	return apc, errors.Wrap(err, "execute action")
}

func handleMessage(c *irc.Client, m *irc.Message, event *string, eventData plugins.FieldCollection) {
	for _, r := range config.GetMatchingRules(m, event, eventData) {
		var preventCooldown bool

		for _, a := range r.Actions {
			apc, err := triggerAction(c, m, r, a, eventData)
			if err != nil {
				log.WithError(err).Error("Unable to trigger action")
				break // Break execution when one action fails
			}
			preventCooldown = preventCooldown || apc
		}

		// Lock command
		if !preventCooldown {
			r.SetCooldown(timerStore, m, eventData)
		}
	}
}

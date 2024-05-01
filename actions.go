package main

import (
	"path"
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/go_helpers/v2/fieldcollection"
	"github.com/Luzifer/twitch-bot/v3/internal/locker"
	"github.com/Luzifer/twitch-bot/v3/plugins"
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

func triggerAction(c *irc.Client, m *irc.Message, rule *plugins.Rule, ra *plugins.RuleAction, eventData *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
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

func handleMessage(c *irc.Client, m *irc.Message, event *string, eventData *fieldcollection.FieldCollection) {
	// Send events to registered handlers
	if event != nil {
		go notifyEventHandlers(*event, eventData)
	}

	matchingRules := config.GetMatchingRules(m, event, eventData)
	for i := range matchingRules {
		go handleMessageRuleExecution(c, m, matchingRules[i], eventData)
	}
}

func handleMessageRuleExecution(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection) {
	locker.LockByKey(path.Join("rule-execution", r.MatcherID()))
	defer locker.UnlockByKey(path.Join("rule-execution", r.MatcherID()))

	var (
		ruleEventData   = fieldcollection.NewFieldCollection()
		preventCooldown bool
	)

	if eventData != nil {
		ruleEventData.SetFromData(eventData.Data())
	}

ActionsLoop:
	for _, a := range r.Actions {
		apc, err := triggerAction(c, m, r, a, ruleEventData)
		switch {
		case err == nil:
			// Rule execution did not cause an error, we store the
			// cooldown modifier and continue

			preventCooldown = preventCooldown || apc
			continue ActionsLoop

		case errors.Is(err, plugins.ErrStopRuleExecution):
			// Action has asked to stop executing this rule so we store
			// the cooldown modifier and stop executing the actions stack
			// Action experienced an error: We don't store the cooldown
			// state of this action and stop executing the actions stack
			// for this rule

			preventCooldown = preventCooldown || apc
			break ActionsLoop

		default:
			// Break execution for this rule when one action fails
			// Lock command

			log.WithError(err).Error("Unable to trigger action")
			break ActionsLoop
		}
	}

	if !preventCooldown {
		r.SetCooldown(timerService, m, eventData)
	}
}

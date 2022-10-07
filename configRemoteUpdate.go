package main

import (
	"fmt"
	"math/rand"

	log "github.com/sirupsen/logrus"
)

func updateConfigCron() string {
	minute := rand.Intn(60) //nolint:gomnd,gosec // Only used to distribute load
	return fmt.Sprintf("%d * * * *", minute)
}

func updateConfigFromRemote() {
	err := patchConfig(
		cfg.Config,
		"Remote Update", "twitch-bot@luzifer.io",
		"update rules from subscription URLs",
		func(cfg *configFile) error {
			var hasUpdate bool

			for _, r := range cfg.Rules {
				logger := log.WithField("rule", r.MatcherID())

				rhu, err := r.UpdateFromSubscription()
				if err != nil {
					logger.WithError(err).Error("updating rule")
					continue
				}

				if rhu {
					hasUpdate = true
					logger.Info("updated rule from remote URL")
				}

			}

			if !hasUpdate {
				return errSaveNotRequired
			}
			return nil
		},
	)
	if err != nil {
		log.WithError(err).Error("updating config rules from subscriptions")
	}
}

package main

import (
	"fmt"
	"math/rand"

	log "github.com/sirupsen/logrus"
)

func updateConfigCron() string {
	minute := rand.Intn(60) //nolint:gomnd,gosec // Only used to distribute load
	return fmt.Sprintf("0 %d * * * *", minute)
}

func updateConfigFromRemote() {
	log.Debug("starting remote rule update")
	err := patchConfig(
		cfg.Config,
		"Remote Update", "twitch-bot@luzifer.io",
		"update rules from subscription URLs",
		func(cfg *configFile) error {
			var updates int

			for _, r := range cfg.Rules {
				logger := log.WithField("rule", r.MatcherID())

				rhu, err := r.UpdateFromSubscription()
				if err != nil {
					logger.WithError(err).Error("updating rule")
					continue
				}

				if rhu {
					updates++
					logger.Info("updated rule from remote URL")
				}
			}

			log.WithField("updates", updates).Debug("remote rule update finished")
			if updates == 0 {
				return errSaveNotRequired
			}
			return nil
		},
	)
	if err != nil {
		log.WithError(err).Error("updating config rules from subscriptions")
	}
}

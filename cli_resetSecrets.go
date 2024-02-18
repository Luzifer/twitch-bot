package main

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func init() {
	cli.Add(cliRegistryEntry{
		Name:        "reset-secrets",
		Description: "Remove encrypted data to reset encryption passphrase",
		Run: func([]string) error {
			if err := accessService.RemoveAllExtendedTwitchCredentials(); err != nil {
				return errors.Wrap(err, "resetting Twitch credentials")
			}
			log.Info("removed stored Twitch credentials")

			if err := db.ResetEncryptedCoreMeta(); err != nil {
				return errors.Wrap(err, "resetting encrypted meta entries")
			}
			log.Info("removed encrypted meta entries")

			return nil
		},
	})
}

package main

import (
	"fmt"

	"github.com/Luzifer/go_helpers/cli"
	log "github.com/sirupsen/logrus"
)

func init() {
	cliTool.Add(cli.RegistryEntry{
		Name:        "reset-secrets",
		Description: "Remove encrypted data to reset encryption passphrase",
		Run: func([]string) error {
			if err := accessService.RemoveAllExtendedTwitchCredentials(); err != nil {
				return fmt.Errorf("resetting Twitch credentials: %w", err)
			}
			log.Info("removed stored Twitch credentials")

			if err := db.ResetEncryptedCoreMeta(); err != nil {
				return fmt.Errorf("resetting encrypted meta entries: %w", err)
			}
			log.Info("removed encrypted meta entries")

			return nil
		},
	})
}

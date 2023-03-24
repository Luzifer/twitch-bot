package main

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v3/internal/v2migrator"
)

func init() {
	cli.Add(cliRegistryEntry{
		Name:        "migrate-v2",
		Description: "Migrate old (*.json.gz) storage file into new database",
		Params:      []string{"<old-file>"},
		Run: func(args []string) error {
			if len(args) < 2 { //nolint:gomnd // Just a count of parameters
				return errors.New("Usage: twitch-bot migrate-v2 <old storage file>")
			}

			v2s := v2migrator.NewStorageFile()
			if err := v2s.Load(args[1], cfg.StorageEncryptionPass); err != nil {
				return errors.Wrap(err, "loading v2 storage file")
			}

			if err := v2s.Migrate(db); err != nil {
				return errors.Wrap(err, "migrating v2 storage file")
			}

			log.Info("v2 storage file was migrated")
			return nil
		},
	})
}

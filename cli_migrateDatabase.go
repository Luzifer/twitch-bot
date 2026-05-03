package main

import (
	"errors"
	"fmt"
	"sync"

	"github.com/Luzifer/go_helpers/cli"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/Luzifer/twitch-bot/v3/pkg/database"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

var (
	dbCopyFuncs     = make(map[string]plugins.DatabaseCopyFunc)
	dbCopyFuncsLock sync.Mutex
)

func init() {
	cliTool.Add(cli.RegistryEntry{
		Name:        "copy-database",
		Description: "Copies database contents to a new storage DSN i.e. for migrating to a new DBMS",
		Params:      []string{"<target storage-type>", "<target DSN>"},
		Run: func(args []string) error {
			if len(args) < 3 { //nolint:mnd // Just a count of parameters
				return errors.New("usage: twitch-bot copy-database <target storage-type> <target DSN>")
			}

			// Core functions cannot register themselves, we take that for them
			registerDatabaseCopyFunc("core-values", db.CopyDatabase)
			registerDatabaseCopyFunc("permissions", accessService.CopyDatabase)
			registerDatabaseCopyFunc("timers", timerService.CopyDatabase)

			targetDB, err := database.New(args[1], args[2], cfg.StorageEncryptionPass)
			if err != nil {
				return fmt.Errorf("connecting to target db: %w", err)
			}
			defer func() {
				if err := targetDB.Close(); err != nil {
					logrus.WithError(err).Error("closing connection to target db")
				}
			}()

			if err := targetDB.DB().Transaction(func(tx *gorm.DB) (err error) {
				for name, dbcf := range dbCopyFuncs {
					logrus.WithField("name", name).Info("running migration")
					if err = dbcf(db.DB(), tx); err != nil {
						return fmt.Errorf("running DatabaseCopyFunc %q: %w", name, err)
					}
				}

				logrus.Info("database has been copied successfully")

				return nil
			}); err != nil {
				return fmt.Errorf("copying database to target: %w", err)
			}

			return nil
		},
	})
}

func registerDatabaseCopyFunc(name string, fn plugins.DatabaseCopyFunc) {
	dbCopyFuncsLock.Lock()
	defer dbCopyFuncsLock.Unlock()

	dbCopyFuncs[name] = fn
}

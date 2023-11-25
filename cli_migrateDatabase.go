package main

import (
	"sync"

	"github.com/Luzifer/twitch-bot/v3/pkg/database"
	"github.com/Luzifer/twitch-bot/v3/plugins"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	dbCopyFuncs     = map[string]plugins.DatabaseCopyFunc{}
	dbCopyFuncsLock sync.Mutex
)

func init() {
	cli.Add(cliRegistryEntry{
		Name:        "copy-database",
		Description: "Copies database contents to a new storage DSN i.e. for migrating to a new DBMS",
		Params:      []string{"<target storage-type>", "<target DSN>"},
		Run: func(args []string) error {
			if len(args) < 3 { //nolint:gomnd // Just a count of parameters
				return errors.New("Usage: twitch-bot copy-database <target storage-type> <target DSN>")
			}

			// Core functions cannot register themselves, we take that for them
			registerDatabaseCopyFunc("core-values", db.CopyDatabase)
			registerDatabaseCopyFunc("permissions", accessService.CopyDatabase)
			registerDatabaseCopyFunc("timers", timerService.CopyDatabase)

			targetDB, err := database.New(args[1], args[2], cfg.StorageEncryptionPass)
			if err != nil {
				return errors.Wrap(err, "connecting to target db")
			}
			defer func() {
				if err := targetDB.Close(); err != nil {
					logrus.WithError(err).Error("closing connection to target db")
				}
			}()

			return errors.Wrap(
				targetDB.DB().Transaction(func(tx *gorm.DB) (err error) {
					for name, dbcf := range dbCopyFuncs {
						logrus.WithField("name", name).Info("running migration")
						if err = dbcf(db.DB(), tx); err != nil {
							return errors.Wrapf(err, "running DatabaseCopyFunc %q", name)
						}
					}

					logrus.Info("database has been copied successfully")

					return nil
				}),
				"copying database to target",
			)
		},
	})
}

func registerDatabaseCopyFunc(name string, fn plugins.DatabaseCopyFunc) {
	dbCopyFuncsLock.Lock()
	defer dbCopyFuncsLock.Unlock()

	dbCopyFuncs[name] = fn
}

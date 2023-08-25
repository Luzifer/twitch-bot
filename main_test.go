package main

import (
	"os"
	"testing"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/twitch-bot/v3/internal/service/access"
	"github.com/Luzifer/twitch-bot/v3/internal/service/timer"
	"github.com/Luzifer/twitch-bot/v3/pkg/database"
)

func TestMain(m *testing.M) {
	var err error

	if db, err = database.New("sqlite", "file::memory:?cache=shared", "encpass"); err != nil {
		log.WithError(err).Fatal("opening storage backend")
	}

	if accessService, err = access.New(db); err != nil {
		log.WithError(err).Fatal("applying access migration")
	}

	cronService = cron.New(cron.WithSeconds())

	if timerService, err = timer.New(db, cronService); err != nil {
		log.WithError(err).Fatal("applying timer migration")
	}

	if err = initCorePlugins(); err != nil {
		log.WithError(err).Fatal("Unable to load core plugins")
	}

	os.Exit(m.Run())
}

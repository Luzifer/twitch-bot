package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/fsnotify.v1"

	"github.com/Luzifer/rconfig/v2"
)

const ircReconnectDelay = 100 * time.Millisecond

var (
	cfg = struct {
		CommandTimeout time.Duration `flag:"command-timeout" default:"30s" description:"Timeout for command execution"`
		Config         string        `flag:"config,c" default:"./config.yaml" description:"Location of configuration file"`
		LogLevel       string        `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`
		StorageFile    string        `flag:"storage-file" default:"./storage.json.gz" description:"Where to store the data"`
		TwitchClient   string        `flag:"twitch-client" default:"" description:"Client ID to act as" validate:"nonzero"`
		TwitchToken    string        `flag:"twitch-token" default:"" description:"OAuth token valid for client"`
		VersionAndExit bool          `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	config     *configFile
	configLock = new(sync.RWMutex)

	store = newStorageFile()

	version = "dev"
)

func init() {
	for _, a := range os.Args {
		if strings.HasPrefix(a, "-test.") {
			// Skip initialize for test run
			return
		}
	}

	rconfig.AutoEnv(true)
	if err := rconfig.ParseAndValidate(&cfg); err != nil {
		log.Fatalf("Unable to parse commandline options: %s", err)
	}

	if cfg.VersionAndExit {
		fmt.Printf("twitch-bot %s\n", version)
		os.Exit(0)
	}

	if l, err := log.ParseLevel(cfg.LogLevel); err != nil {
		log.WithError(err).Fatal("Unable to parse log level")
	} else {
		log.SetLevel(l)
	}
}

//nolint: gocognit,gocyclo // Complexity is a little too high but makes no sense to split
func main() {
	var err error

	if err = store.Load(); err != nil {
		log.WithError(err).Fatal("Unable to load storage file")
	}

	if err = loadConfig(cfg.Config); err != nil {
		log.WithError(err).Fatal("Initial config load failed")
	}

	fswatch, err := fsnotify.NewWatcher()
	if err != nil {
		log.WithError(err).Fatal("Unable to create file watcher")
	}

	if err = fswatch.Add(cfg.Config); err != nil {
		log.WithError(err).Error("Unable to watch config, auto-reload will not work")
	}

	var (
		irc               *ircHandler
		ircDisconnected   = make(chan struct{}, 1)
		autoMessageTicker = time.NewTicker(time.Second)
	)

	ircDisconnected <- struct{}{}

	for {
		select {

		case <-ircDisconnected:
			if irc != nil {
				irc.Close()
			}

			if irc, err = newIRCHandler(); err != nil {
				log.WithError(err).Fatal("Unable to create IRC client")
			}

			go func() {
				if err := irc.Run(); err != nil {
					log.WithError(err).Error("IRC run exited unexpectedly")
				}
				time.Sleep(ircReconnectDelay)
				ircDisconnected <- struct{}{}
			}()

		case evt := <-fswatch.Events:
			if evt.Op&fsnotify.Write != fsnotify.Write {
				continue
			}

			if err := loadConfig(cfg.Config); err != nil {
				log.WithError(err).Error("Unable to reload config")
				continue
			}

			irc.ExecuteJoins(config.Channels)

			log.Info("Config file reloaded")

		case <-autoMessageTicker.C:
			configLock.RLock()
			for _, am := range config.AutoMessages {
				if !am.CanSend() {
					continue
				}

				if err := am.Send(irc.c); err != nil {
					log.WithError(err).Error("Unable to send automated message")
				}
			}
			configLock.RUnlock()

		}
	}
}

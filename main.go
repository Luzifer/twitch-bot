package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-irc/irc"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/Luzifer/rconfig/v2"
	"github.com/Luzifer/twitch-bot/twitch"
)

const ircReconnectDelay = 100 * time.Millisecond

var (
	cfg = struct {
		CommandTimeout time.Duration `flag:"command-timeout" default:"30s" description:"Timeout for command execution"`
		Config         string        `flag:"config,c" default:"./config.yaml" description:"Location of configuration file"`
		IRCRateLimit   time.Duration `flag:"rate-limit" default:"1500ms" description:"How often to send a message (default: 20/30s=1500ms, if your bot is mod everywhere: 100/30s=300ms, different for known/verified bots)"`
		LogLevel       string        `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`
		PluginDir      string        `flag:"plugin-dir" default:"/usr/lib/twitch-bot" description:"Where to find and load plugins"`
		StorageFile    string        `flag:"storage-file" default:"./storage.json.gz" description:"Where to store the data"`
		TwitchClient   string        `flag:"twitch-client" default:"" description:"Client ID to act as"`
		TwitchToken    string        `flag:"twitch-token" default:"" description:"OAuth token valid for client"`
		ValidateConfig bool          `flag:"validate-config,v" default:"false" description:"Loads the config, logs any errors and quits with status 0 on success"`
		VersionAndExit bool          `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	config     *configFile
	configLock = new(sync.RWMutex)

	cronService *cron.Cron
	ircHdl      *ircHandler
	router      = mux.NewRouter()

	sendMessage func(m *irc.Message) error

	store        = newStorageFile(false)
	twitchClient *twitch.Client

	version = "dev"
)

func init() {
	for _, a := range os.Args {
		if strings.HasPrefix(a, "-test.") {
			// Skip initialize for test run
			store = newStorageFile(true) // Use in-mem-store for tests
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

//nolint: funlen,gocognit,gocyclo // Complexity is a little too high but makes no sense to split
func main() {
	var err error

	cronService = cron.New()
	twitchClient = twitch.New(cfg.TwitchClient, cfg.TwitchToken)

	twitchWatch := newTwitchWatcher()
	cronService.AddFunc("@every 10s", twitchWatch.Check) // Query may run that often as the twitchClient has an internal cache

	router.HandleFunc("/", handleSwaggerHTML)
	router.HandleFunc("/openapi.json", handleSwaggerRequest)

	if err = loadPlugins(cfg.PluginDir); err != nil {
		log.WithError(err).Fatal("Unable to load plugins")
	}

	if err = loadConfig(cfg.Config); err != nil {
		log.WithError(err).Fatal("Initial config load failed")
	}
	defer func() { config.CloseRawMessageWriter() }()

	if cfg.ValidateConfig {
		// We were asked to only validate the config, this was successful
		log.Info("Config validated successfully")
		return
	}

	for _, c := range config.Channels {
		if err := twitchWatch.AddChannel(c); err != nil {
			log.WithError(err).WithField("channel", c).Error("Unable to add channel to watcher")
		}
	}

	if err = startCheck(); err != nil {
		log.WithError(err).Fatal("Missing required parameters")
	}

	if err = store.Load(); err != nil {
		log.WithError(err).Fatal("Unable to load storage file")
	}

	fsEvents := make(chan configChangeEvent, 1)
	go watchConfigChanges(cfg.Config, fsEvents)

	var (
		ircDisconnected   = make(chan struct{}, 1)
		autoMessageTicker = time.NewTicker(time.Second)
	)

	cronService.Start()

	if config.HTTPListen != "" {
		// If listen address is configured start HTTP server
		go http.ListenAndServe(config.HTTPListen, router)
	}

	ircDisconnected <- struct{}{}

	for {
		select {

		case <-ircDisconnected:
			if ircHdl != nil {
				sendMessage = nil
				ircHdl.Close()
			}

			if ircHdl, err = newIRCHandler(); err != nil {
				log.WithError(err).Fatal("Unable to create IRC client")
			}

			go func() {
				sendMessage = ircHdl.SendMessage
				if err := ircHdl.Run(); err != nil {
					log.WithError(err).Error("IRC run exited unexpectedly")
				}
				sendMessage = nil
				time.Sleep(ircReconnectDelay)
				ircDisconnected <- struct{}{}
			}()

		case evt := <-fsEvents:
			switch evt {
			case configChangeEventUnkown:
				continue

			case configChangeEventNotExist:
				log.Error("Config file is not available, not reloading config")
				continue

			case configChangeEventModified:
				// Fine, reload
			}

			previousChannels := append([]string{}, config.Channels...)

			if err := loadConfig(cfg.Config); err != nil {
				log.WithError(err).Error("Unable to reload config")
				continue
			}

			ircHdl.ExecuteJoins(config.Channels)
			for _, c := range config.Channels {
				if err := twitchWatch.AddChannel(c); err != nil {
					log.WithError(err).WithField("channel", c).Error("Unable to add channel to watcher")
				}
			}

			for _, c := range previousChannels {
				if !str.StringInSlice(c, config.Channels) {
					log.WithField("channel", c).Info("Leaving removed channel...")
					ircHdl.ExecutePart(c)

					if err := twitchWatch.RemoveChannel(c); err != nil {
						log.WithError(err).WithField("channel", c).Error("Unable to remove channel from watcher")
					}
				}
			}

		case <-autoMessageTicker.C:
			configLock.RLock()
			for _, am := range config.AutoMessages {
				if !am.CanSend() {
					continue
				}

				if err := am.Send(ircHdl.c); err != nil {
					log.WithError(err).Error("Unable to send automated message")
				}
			}
			configLock.RUnlock()

		}
	}
}

func startCheck() error {
	var errs []string

	if cfg.TwitchClient == "" {
		errs = append(errs, "No Twitch-ClientId given")
	}

	if cfg.TwitchToken == "" {
		errs = append(errs, "Twitch-Token is unset")
	}

	if len(errs) > 0 {
		fmt.Println(`
You've not provided a Twitch-ClientId and/or a Twitch-Token.

These parameters are required and you need to provide them. In case
you need help with obtaining those credentials please visit the
following website:

         https://luzifer.github.io/twitch-bot/

You will be guided through the token generation and can afterwards
provide the required configuration parameters.`)
		return errors.New(strings.Join(errs, ", "))
	}

	return nil
}

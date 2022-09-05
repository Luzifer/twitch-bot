package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid/v3"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/Luzifer/rconfig/v2"
	"github.com/Luzifer/twitch-bot/twitch"
)

const (
	ircReconnectDelay = 100 * time.Millisecond

	initialIRCRetryBackoff    = 500 * time.Millisecond
	ircRetryBackoffMultiplier = 1.5
	maxIRCRetryBackoff        = time.Minute
)

var (
	cfg = struct {
		BaseURL               string        `flag:"base-url" default:"" description:"External URL of the config-editor interface (set to enable EventSub support)"`
		CommandTimeout        time.Duration `flag:"command-timeout" default:"30s" description:"Timeout for command execution"`
		Config                string        `flag:"config,c" default:"./config.yaml" description:"Location of configuration file"`
		IRCRateLimit          time.Duration `flag:"rate-limit" default:"1500ms" description:"How often to send a message (default: 20/30s=1500ms, if your bot is mod everywhere: 100/30s=300ms, different for known/verified bots)"`
		LogLevel              string        `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`
		PluginDir             string        `flag:"plugin-dir" default:"/usr/lib/twitch-bot" description:"Where to find and load plugins"`
		StorageFile           string        `flag:"storage-file" default:"./storage.json.gz" description:"Where to store the data"`
		StorageEncryptionPass string        `flag:"storage-encryption-pass" default:"" description:"Passphrase to encrypt secrets inside storage (defaults to twitch-client:twitch-client-secret)"`
		TwitchClient          string        `flag:"twitch-client" default:"" description:"Client ID to act as"`
		TwitchClientSecret    string        `flag:"twitch-client-secret" default:"" description:"Secret for the Client ID"`
		TwitchToken           string        `flag:"twitch-token" default:"" description:"OAuth token valid for client (fallback if no token was set in interface)"`
		ValidateConfig        bool          `flag:"validate-config,v" default:"false" description:"Loads the config, logs any errors and quits with status 0 on success"`
		VersionAndExit        bool          `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	config     *configFile
	configLock = new(sync.RWMutex)

	botUserstate = newTwitchUserStateStore()
	cronService  *cron.Cron
	ircHdl       *ircHandler
	router       = mux.NewRouter()

	runID                 = uuid.Must(uuid.NewV4()).String()
	externalHTTPAvailable bool

	store                = newStorageFile(false)
	twitchClient         *twitch.Client
	twitchEventSubClient *twitch.EventSubClient

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

	if cfg.StorageEncryptionPass == "" {
		log.Warn("No storage encryption passphrase was set, falling back to client-id:client-secret")
		cfg.StorageEncryptionPass = strings.Join([]string{
			cfg.TwitchClient,
			cfg.TwitchClientSecret,
		}, ":")
	}
}

func handleSubCommand(args []string) {
	switch args[0] {

	case "actor-docs":
		doc, err := generateActorDocs()
		if err != nil {
			log.WithError(err).Fatal("Unable to generate actor docs")
		}
		if _, err = os.Stdout.Write(append(bytes.TrimSpace(doc), '\n')); err != nil {
			log.WithError(err).Fatal("Unable to write actor docs to stdout")
		}

	case "api-token":
		if len(args) < 3 { //nolint:gomnd // Just a count of parameters
			log.Fatalf("Usage: twitch-bot api-token <token name> <scope> [...scope]")
		}

		t := configAuthToken{
			Name:    args[1],
			Modules: args[2:],
		}

		if err := fillAuthToken(&t); err != nil {
			log.WithError(err).Fatal("Unable to generate token")
		}

		log.WithField("token", t.Token).Info("Token generated, add this to your config:")
		if err := yaml.NewEncoder(os.Stdout).Encode(map[string]map[string]configAuthToken{
			"auth_tokens": {
				uuid.Must(uuid.NewV4()).String(): t,
			},
		}); err != nil {
			log.WithError(err).Fatal("Unable to output token info")
		}

	case "help":
		fmt.Println("Supported sub-commands are:")
		fmt.Println("  actor-docs                     Generate markdown documentation for available actors")
		fmt.Println("  api-token <name> <scope...>    Generate an api-token to be entered into the config")
		fmt.Println("  help                           Prints this help message")

	default:
		handleSubCommand([]string{"help"})
		log.Fatalf("Unknown sub-command %q", args[0])

	}
}

//nolint: funlen,gocognit,gocyclo // Complexity is a little too high but makes no sense to split
func main() {
	var err error

	if err = store.Load(); err != nil {
		log.WithError(err).Fatal("Unable to load storage file")
	}

	cronService = cron.New()
	twitchClient = twitch.New(cfg.TwitchClient, cfg.TwitchClientSecret, store.GetBotToken(cfg.TwitchToken), store.BotRefreshToken)
	twitchClient.SetTokenUpdateHook(func(at, rt string) error {
		if err := store.UpdateBotToken(at, rt); err != nil {
			return errors.Wrap(err, "updating store")
		}

		// Misuse the config reload hook to let the frontend reload its state
		configReloadHooksLock.RLock()
		defer configReloadHooksLock.RUnlock()
		for _, fn := range configReloadHooks {
			fn()
		}

		return nil
	})

	twitchWatch := newTwitchWatcher()
	// Query may run that often as the twitchClient has an internal
	// cache but shouldn't run more often as EventSub subscriptions
	// are retried on error each time
	cronService.AddFunc("@every 30s", twitchWatch.Check)

	router.Use(corsMiddleware)
	router.HandleFunc("/openapi.html", handleSwaggerHTML)
	router.HandleFunc("/openapi.json", handleSwaggerRequest)
	router.HandleFunc("/selfcheck", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(runID)) })

	router.MethodNotAllowedHandler = corsMiddleware(http.HandlerFunc(func(res http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			// Most likely JS client asking for CORS headers
			res.WriteHeader(http.StatusNoContent)
			return
		}

		res.WriteHeader(http.StatusMethodNotAllowed)
	}))

	if err = initCorePlugins(); err != nil {
		log.WithError(err).Fatal("Unable to load core plugins")
	}

	if err = loadPlugins(cfg.PluginDir); err != nil {
		log.WithError(err).Fatal("Unable to load plugins")
	}

	if len(rconfig.Args()) > 1 {
		handleSubCommand(rconfig.Args()[1:])
		return
	}

	if err = loadConfig(cfg.Config); err != nil {
		if os.IsNotExist(errors.Cause(err)) {
			if err = writeDefaultConfigFile(cfg.Config); err != nil {
				log.WithError(err).Fatal("Initial config not found and not able to create example config")
			}

			log.WithField("filename", cfg.Config).Warn("No config was found, created example config: Please review that config!")
			return
		}

		log.WithError(err).Fatal("Initial config load failed")
	}
	defer func() { config.CloseRawMessageWriter() }()

	if cfg.ValidateConfig {
		// We were asked to only validate the config, this was successful
		log.Info("Config validated successfully")
		return
	}

	if err = startCheck(); err != nil {
		log.WithError(err).Fatal("Missing required parameters")
	}

	fsEvents := make(chan configChangeEvent, 1)
	go watchConfigChanges(cfg.Config, fsEvents)

	var (
		ircDisconnected   = make(chan struct{}, 1)
		ircRetryBackoff   = initialIRCRetryBackoff
		autoMessageTicker = time.NewTicker(time.Second)
	)

	cronService.Start()

	if config.HTTPListen != "" {
		// If listen address is configured start HTTP server
		listener, err := net.Listen("tcp", config.HTTPListen)
		if err != nil {
			log.WithError(err).Fatal("Unable to open http_listen port")
		}

		go http.Serve(listener, router)
		log.WithField("address", listener.Addr().String()).Info("HTTP server started")

		checkExternalHTTP()

		if externalHTTPAvailable && cfg.TwitchClient != "" && cfg.TwitchClientSecret != "" {
			secret, handle, err := store.GetOrGenerateEventSubSecret()
			if err != nil {
				log.WithError(err).Fatal("Unable to get or create eventsub secret")
			}

			twitchEventSubClient, err = twitch.NewEventSubClient(twitchClient, strings.Join([]string{
				strings.TrimRight(cfg.BaseURL, "/"),
				"eventsub",
			}, "/"), secret, handle)

			if err != nil {
				log.WithError(err).Fatal("Unable to create eventsub client")
			}

			if err := twitchWatch.registerGlobalHooks(); err != nil {
				log.WithError(err).Fatal("Unable to register global eventsub hooks")
			}

			router.HandleFunc("/eventsub/{keyhandle}", twitchEventSubClient.HandleEventsubPush).Methods(http.MethodPost)
		}
	}

	for _, c := range config.Channels {
		if err := twitchWatch.AddChannel(c); err != nil {
			log.WithError(err).WithField("channel", c).Error("Unable to add channel to watcher")
		}
	}

	ircDisconnected <- struct{}{}

	for {
		select {

		case <-ircDisconnected:
			if ircHdl != nil {
				ircHdl.Close()
			}

			if ircHdl, err = newIRCHandler(); err != nil {
				log.WithError(err).Error("Unable to connect to IRC")
				go func() {
					time.Sleep(ircRetryBackoff)
					ircRetryBackoff = time.Duration(math.Min(float64(maxIRCRetryBackoff), float64(ircRetryBackoff)*ircRetryBackoffMultiplier))
					ircDisconnected <- struct{}{}
				}()
				continue
			}

			ircRetryBackoff = initialIRCRetryBackoff // Successfully created, reset backoff

			go func() {
				if err := ircHdl.Run(); err != nil {
					log.WithError(err).Error("IRC run exited unexpectedly")
				}
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

func checkExternalHTTP() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	base, err := url.Parse(cfg.BaseURL)
	if err != nil {
		log.WithError(err).Error("Unable to parse BaseURL")
		return
	}

	if base.String() == "" {
		log.Debug("No BaseURL set, disabling EventSub support")
		return
	}

	base.Path = strings.Join([]string{
		strings.TrimRight(base.Path, "/"),
		"selfcheck",
	}, "/")

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.WithError(err).Error("Unable to fetch selfcheck")
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Error("Unable to read selfcheck response")
		return
	}

	if strings.TrimSpace(string(data)) == runID {
		externalHTTPAvailable = true
		log.Debug("Self-Check successful, EventSub support is available")
	} else {
		externalHTTPAvailable = false
		log.Debug("Self-Check failed, EventSub support is not available")
	}
}

func startCheck() error {
	var errs []string

	if cfg.TwitchClient == "" {
		errs = append(errs, "No Twitch-ClientId given")
	}

	if cfg.TwitchClientSecret == "" {
		errs = append(errs, "No Twitch-ClientSecret given")
	}

	if len(errs) > 0 {
		fmt.Println(`
You've not provided a Twitch-ClientId and/or a Twitch-ClientSecret.

These parameters are required and you need to provide them.

The Twitch Token can be set through the web-interface. In case you
want to set it through parameters and need help with obtaining it,
please visit the following website:

         https://luzifer.github.io/twitch-bot/

You will be guided through the token generation and can afterwards
provide the required configuration parameters.`)
		return errors.New(strings.Join(errs, ", "))
	}

	return nil
}

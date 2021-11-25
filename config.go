package main

import (
	_ "embed"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	"github.com/go-irc/irc"
	"github.com/gofrs/uuid/v3"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/Luzifer/twitch-bot/plugins"
)

const expectedMinConfigVersion = 2

var (
	//go:embed default_config.yaml
	defaultConfigurationYAML []byte

	hashstructUUIDNamespace = uuid.Must(uuid.FromString("3a0ccc46-d3ba-46b5-ac07-27528c933174"))

	configReloadHooks     = map[string]func(){}
	configReloadHooksLock sync.RWMutex
)

func registerConfigReloadHook(hook func()) func() {
	configReloadHooksLock.Lock()
	defer configReloadHooksLock.Unlock()

	id := uuid.Must(uuid.NewV4()).String()
	configReloadHooks[id] = hook

	return func() {
		configReloadHooksLock.Lock()
		defer configReloadHooksLock.Unlock()

		delete(configReloadHooks, id)
	}
}

type (
	configAuthToken struct {
		Hash    string   `json:"-" yaml:"hash"`
		Modules []string `json:"modules" yaml:"modules"`
		Name    string   `json:"name" yaml:"name"`
		Token   string   `json:"token" yaml:"-"`
	}

	configFileVersioner struct {
		ConfigVersion int64 `yaml:"config_version"`
	}

	configFile struct {
		AuthTokens           map[string]configAuthToken `yaml:"auth_tokens"`
		AutoMessages         []*autoMessage             `yaml:"auto_messages"`
		BotEditors           []string                   `yaml:"bot_editors"`
		Channels             []string                   `yaml:"channels"`
		GitTrackConfig       bool                       `yaml:"git_track_config"`
		HTTPListen           string                     `yaml:"http_listen"`
		PermitAllowModerator bool                       `yaml:"permit_allow_moderator"`
		PermitTimeout        time.Duration              `yaml:"permit_timeout"`
		RawLog               string                     `yaml:"raw_log"`
		Rules                []*plugins.Rule            `yaml:"rules"`
		Variables            map[string]interface{}     `yaml:"variables"`

		rawLogWriter io.WriteCloser

		configFileVersioner `yaml:",inline"`
	}
)

func newConfigFile() *configFile {
	return &configFile{
		AuthTokens:    map[string]configAuthToken{},
		PermitTimeout: time.Minute,
	}
}

func loadConfig(filename string) error {
	var (
		configVersion = &configFileVersioner{}
		err           error
		tmpConfig     = newConfigFile()
	)

	if err = parseConfigFromYAML(filename, configVersion, false); err != nil {
		return errors.Wrap(err, "parsing config version")
	}

	if configVersion.ConfigVersion < expectedMinConfigVersion {
		return errors.Errorf("config version too old: %d < %d - Please have a look at the documentation!", configVersion.ConfigVersion, expectedMinConfigVersion)
	}

	if err = parseConfigFromYAML(filename, tmpConfig, true); err != nil {
		return errors.Wrap(err, "parsing config")
	}

	if len(tmpConfig.Channels) == 0 {
		log.Warn("Loaded config with empty channel list")
	}

	if len(tmpConfig.Rules) == 0 {
		log.Warn("Loaded config with empty ruleset")
	}

	if err = tmpConfig.validateRuleActions(); err != nil {
		return errors.Wrap(err, "validating rule actions")
	}

	configLock.Lock()
	defer configLock.Unlock()

	tmpConfig.updateAutoMessagesFromConfig(config)
	tmpConfig.fixDurations()
	tmpConfig.fixMissingUUIDs()

	switch {
	case config != nil && config.RawLog == tmpConfig.RawLog:
		tmpConfig.rawLogWriter = config.rawLogWriter

	case tmpConfig.RawLog == "":
		if err = config.CloseRawMessageWriter(); err != nil {
			return errors.Wrap(err, "closing old raw log writer")
		}

		tmpConfig.rawLogWriter = writeNoOpCloser{io.Discard}

	default:
		if err = config.CloseRawMessageWriter(); err != nil {
			return errors.Wrap(err, "closing old raw log writer")
		}
		if err = os.MkdirAll(path.Dir(tmpConfig.RawLog), 0o755); err != nil { //nolint:gomnd // This is a common directory permission
			return errors.Wrap(err, "creating directories for raw log")
		}
		if tmpConfig.rawLogWriter, err = os.OpenFile(tmpConfig.RawLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err != nil { //nolint:gomnd // This is a common file permission
			return errors.Wrap(err, "opening raw log for appending")
		}
	}

	config = tmpConfig

	log.WithFields(log.Fields{
		"auto_messages": len(config.AutoMessages),
		"rules":         len(config.Rules),
		"channels":      len(config.Channels),
	}).Info("Config file (re)loaded")

	// Notify listener config has changed
	configReloadHooksLock.RLock()
	defer configReloadHooksLock.RUnlock()
	for _, fn := range configReloadHooks {
		fn()
	}

	return nil
}

func parseConfigFromYAML(filename string, obj interface{}, strict bool) error {
	f, err := os.Open(filename)
	if err != nil {
		return errors.Wrap(err, "open config file")
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	decoder.SetStrict(strict)

	return errors.Wrap(decoder.Decode(obj), "decode config file")
}

func patchConfig(filename, authorName, authorEmail, summary string, patcher func(*configFile) error) error {
	var (
		cfgFile = newConfigFile()
		err     error
	)

	if err = parseConfigFromYAML(filename, cfgFile, true); err != nil {
		return errors.Wrap(err, "loading current config")
	}

	cfgFile.fixMissingUUIDs()

	if err = patcher(cfgFile); err != nil {
		return errors.Wrap(err, "patching config")
	}

	return errors.Wrap(
		writeConfigToYAML(filename, authorName, authorEmail, summary, cfgFile),
		"replacing config",
	)
}

func writeConfigToYAML(filename, authorName, authorEmail, summary string, obj *configFile) error {
	tmpFile, err := ioutil.TempFile(path.Dir(filename), "twitch-bot-*.yaml")
	if err != nil {
		return errors.Wrap(err, "opening tempfile")
	}
	tmpFileName := tmpFile.Name()

	fmt.Fprintf(tmpFile, "# Automatically updated by %s using Config-Editor frontend, last update: %s\n", authorName, time.Now().Format(time.RFC3339))

	if err = yaml.NewEncoder(tmpFile).Encode(obj); err != nil {
		tmpFile.Close()
		return errors.Wrap(err, "encoding config")
	}
	tmpFile.Close()

	if err = os.Rename(tmpFileName, filename); err != nil {
		return errors.Wrap(err, "moving config to location")
	}

	if !obj.GitTrackConfig {
		return nil
	}

	git := newGitHelper(path.Dir(filename))
	if !git.HasRepo() {
		log.Error("Instructed to track changes using Git, but config not in repo")
		return nil
	}

	return errors.Wrap(
		git.CommitChange(path.Base(filename), authorName, authorEmail, summary),
		"committing config changes",
	)
}

func writeDefaultConfigFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return errors.Wrap(err, "creating config file")
	}
	defer f.Close()

	_, err = f.Write(defaultConfigurationYAML)
	return errors.Wrap(err, "writing default config")
}

func (c *configFile) CloseRawMessageWriter() error {
	if c == nil || c.rawLogWriter == nil {
		return nil
	}
	return c.rawLogWriter.Close()
}

func (c configFile) GetMatchingRules(m *irc.Message, event *string, eventData *plugins.FieldCollection) []*plugins.Rule {
	configLock.RLock()
	defer configLock.RUnlock()

	var out []*plugins.Rule

	for _, r := range c.Rules {
		if r.Matches(m, event, timerStore, formatMessage, twitchClient, eventData) {
			out = append(out, r)
		}
	}

	return out
}

func (c configFile) LogRawMessage(m *irc.Message) error {
	_, err := fmt.Fprintln(c.rawLogWriter, m.String())
	return errors.Wrap(err, "writing raw log message")
}

func (c *configFile) fixDurations() {
	// General fields
	c.PermitTimeout = c.fixedDuration(c.PermitTimeout)

	// Fix rules
	for _, r := range c.Rules {
		r.Cooldown = c.fixedDurationPtr(r.Cooldown)
	}
}

func (configFile) fixedDuration(d time.Duration) time.Duration {
	if d > time.Second {
		return d
	}
	return d * time.Second
}

func (configFile) fixedDurationPtr(d *time.Duration) *time.Duration {
	if d == nil || *d > time.Second {
		return d
	}
	fd := *d * time.Second
	return &fd
}

func (c *configFile) fixMissingUUIDs() {
	for i := range c.AutoMessages {
		if c.AutoMessages[i].UUID != "" {
			continue
		}
		c.AutoMessages[i].UUID = uuid.NewV5(hashstructUUIDNamespace, c.AutoMessages[i].ID()).String()
	}

	for i := range c.Rules {
		if c.Rules[i].UUID != "" {
			continue
		}
		c.Rules[i].UUID = uuid.NewV5(hashstructUUIDNamespace, c.Rules[i].MatcherID()).String()
	}
}

func (c *configFile) updateAutoMessagesFromConfig(old *configFile) {
	for idx, nam := range c.AutoMessages {
		// By default assume last message to be sent now
		// in order not to spam messages at startup
		nam.lastMessageSent = time.Now()

		if !nam.IsValid() {
			log.WithField("index", idx).Warn("Auto-Message configuration is invalid and therefore disabled")
		}

		if old == nil {
			// Initial config load, do not update timers
			continue
		}

		for _, oam := range old.AutoMessages {
			if nam.ID() != oam.ID() {
				continue
			}

			// We disable the old message as executing it would
			// mess up the constraints of the new message
			oam.lock.Lock()
			oam.disabled = true

			nam.lastMessageSent = oam.lastMessageSent
			nam.linesSinceLastMessage = oam.linesSinceLastMessage
			oam.lock.Unlock()
		}
	}
}

func (c configFile) validateRuleActions() error {
	for _, r := range c.Rules {
		logger := log.WithField("rule", r.MatcherID())
		for idx, a := range r.Actions {
			actor, err := getActorByName(a.Type)
			if err != nil {
				logger.WithField("index", idx).WithError(err).Error("Cannot get actor by type")
				return errors.Wrap(err, "getting actor by type")
			}

			if err = actor.Validate(a.Attributes); err != nil {
				logger.WithField("index", idx).WithError(err).Error("Actor reported invalid config")
				return errors.Wrap(err, "validating action attributes")
			}
		}
	}

	return nil
}

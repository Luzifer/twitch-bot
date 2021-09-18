package main

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const expectedMinConfigVersion = 2

//go:embed default_config.yaml
var defaultConfigurationYAML []byte

type (
	configFileVersioner struct {
		ConfigVersion int64 `yaml:"config_version"`
	}

	configFile struct {
		AutoMessages         []*autoMessage         `yaml:"auto_messages"`
		Channels             []string               `yaml:"channels"`
		HTTPListen           string                 `yaml:"http_listen"`
		PermitAllowModerator bool                   `yaml:"permit_allow_moderator"`
		PermitTimeout        time.Duration          `yaml:"permit_timeout"`
		RawLog               string                 `yaml:"raw_log"`
		Rules                []*plugins.Rule        `yaml:"rules"`
		Variables            map[string]interface{} `yaml:"variables"`

		rawLogWriter io.WriteCloser

		configFileVersioner `yaml:",inline"`
	}
)

func newConfigFile() *configFile {
	return &configFile{
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

func (c configFile) GetMatchingRules(m *irc.Message, event *string, eventData map[string]interface{}) []*plugins.Rule {
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

	// Fix auto-messages
	for _, a := range c.AutoMessages {
		a.TimeInterval = c.fixedDuration(a.TimeInterval)
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

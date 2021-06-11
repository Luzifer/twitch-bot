package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type configFile struct {
	AutoMessages         []*autoMessage         `yaml:"auto_messages"`
	Channels             []string               `yaml:"channels"`
	PermitAllowModerator bool                   `yaml:"permit_allow_moderator"`
	PermitTimeout        time.Duration          `yaml:"permit_timeout"`
	RawLog               string                 `yaml:"raw_log"`
	Rules                []*Rule                `yaml:"rules"`
	Variables            map[string]interface{} `yaml:"variables"`

	rawLogWriter io.WriteCloser
}

func newConfigFile() *configFile {
	return &configFile{
		PermitTimeout: time.Minute,
	}
}

func loadConfig(filename string) error {
	var (
		err       error
		tmpConfig *configFile
	)

	switch path.Ext(filename) {
	case ".yaml", ".yml":
		tmpConfig, err = parseConfigFromYAML(filename)

	default:
		return errors.Errorf("Unknown config format %q", path.Ext(filename))
	}

	if err != nil {
		return errors.Wrap(err, "parsing config")
	}

	if len(tmpConfig.Channels) == 0 {
		log.Warn("Loaded config with empty channel list")
	}

	if len(tmpConfig.Rules) == 0 {
		log.Warn("Loaded config with empty ruleset")
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
		if err = os.MkdirAll(path.Dir(tmpConfig.RawLog), 0o755); err != nil {
			return errors.Wrap(err, "creating directories for raw log")
		}
		if tmpConfig.rawLogWriter, err = os.OpenFile(tmpConfig.RawLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err != nil {
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

func parseConfigFromYAML(filename string) (*configFile, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "open config file")
	}
	defer f.Close()

	var (
		decoder   = yaml.NewDecoder(f)
		tmpConfig = newConfigFile()
	)

	decoder.SetStrict(true)

	if err = decoder.Decode(&tmpConfig); err != nil {
		return nil, errors.Wrap(err, "decode config file")
	}

	return tmpConfig, nil
}

func (c *configFile) CloseRawMessageWriter() error {
	if c == nil || c.rawLogWriter == nil {
		return nil
	}
	return c.rawLogWriter.Close()
}

func (c configFile) GetMatchingRules(m *irc.Message, event *string) []*Rule {
	configLock.RLock()
	defer configLock.RUnlock()

	var out []*Rule

	for _, r := range c.Rules {
		if r.matches(m, event) {
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

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
	AutoMessages         []*autoMessage `yaml:"auto_messages"`
	Channels             []string       `yaml:"channels"`
	PermitAllowModerator bool           `yaml:"permit_allow_moderator"`
	PermitTimeout        time.Duration  `yaml:"permit_timeout"`
	RawLog               string         `yaml:"raw_log"`
	Rules                []*rule        `yaml:"rules"`

	rawLogWriter io.WriteCloser
}

func newConfigFile() *configFile {
	return &configFile{
		PermitTimeout: time.Minute,
	}
}

func loadConfig(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return errors.Wrap(err, "open config file")
	}
	defer f.Close()

	var (
		decoder   = yaml.NewDecoder(f)
		tmpConfig = newConfigFile()
	)

	decoder.SetStrict(true)

	if err = decoder.Decode(&tmpConfig); err != nil {
		return errors.Wrap(err, "decode config file")
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

func (c *configFile) CloseRawMessageWriter() error {
	if c == nil || c.rawLogWriter == nil {
		return nil
	}
	return c.rawLogWriter.Close()
}

func (c configFile) GetMatchingRules(m *irc.Message, event *string) []*rule {
	configLock.RLock()
	defer configLock.RUnlock()

	var out []*rule

	for _, r := range c.Rules {
		if r.Matches(m, event) {
			out = append(out, r)
		}
	}

	return out
}

func (c configFile) LogRawMessage(m *irc.Message) error {
	_, err := fmt.Fprintln(c.rawLogWriter, m.String())
	return errors.Wrap(err, "writing raw log message")
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
		}
	}
}

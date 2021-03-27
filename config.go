package main

import (
	"os"
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
	Rules                []*rule        `yaml:"rules"`
}

func newConfigFile() configFile {
	return configFile{
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

	for idx, nam := range tmpConfig.AutoMessages {
		// By default assume last message to be sent now
		// in order not to spam messages at startup
		nam.lastMessageSent = time.Now()

		if !nam.IsValid() {
			log.WithField("index", idx).Warn("Auto-Message configuration is invalid and therefore disabled")
		}

		if config == nil {
			// Initial config load, do not update timers
			continue
		}

		for _, oam := range config.AutoMessages {
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

	configLock.Lock()
	defer configLock.Unlock()

	config = &tmpConfig
	return nil
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

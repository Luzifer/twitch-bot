package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type configFile struct {
	Channels             []string      `yaml:"channels"`
	PermitAllowModerator bool          `yaml:"permit_allow_moderator"`
	PermitTimeout        time.Duration `yaml:"permit_timeout"`
	Rules                []*rule       `yaml:"rules"`
}

func newConfigFile() configFile {
	return configFile{
		PermitTimeout: time.Minute,
	}
}

type rule struct {
	Actions []*ruleAction `yaml:"actions"`

	Cooldown *time.Duration `yaml:"cooldown"`

	MatchChannel *string `yaml:"match_channel"`
	MatchEvent   *string `yaml:"match_event"`
	MatchMessage *string `yaml:"match_message"`

	DisableOnPermit bool     `yaml:"disable_on_permit"`
	DisableOn       []string `yaml:"disable_on"`
	EnableOn        []string `yaml:"enable_on"`

	matchMessage *regexp.Regexp
}

func (r rule) MatcherID() string {
	out := sha256.New()

	for _, e := range []*string{
		r.MatchChannel,
		r.MatchEvent,
		r.MatchMessage,
	} {
		if e != nil {
			fmt.Fprintf(out, *e)
		}
	}

	return fmt.Sprintf("sha256:%x", out.Sum(nil))
}

func (r rule) Matches(m *irc.Message, event *string) bool {
	var err error

	var (
		badges = ircHandler{}.ParseBadgeLevels(m)
		logger = log.WithFields(log.Fields{
			"msg":  m,
			"rule": r,
		})
	)

	// Check Channel match
	if r.MatchChannel != nil {
		if len(m.Params) == 0 || (m.Params[0] != *r.MatchChannel) {
			logger.Trace("Non-Match: Channel")
			return false
		}
	}

	// Check Event match
	if r.MatchEvent != nil {
		if event == nil || *r.MatchEvent != *event {
			logger.Trace("Non-Match: Event")
			return false
		}
	}

	// Check Message match
	if r.MatchMessage != nil {
		// If the regexp was not yet compiled, cache it
		if r.matchMessage == nil {
			if r.matchMessage, err = regexp.Compile(*r.MatchMessage); err != nil {
				logger.WithError(err).Error("Unable to compile expression")
				return false
			}
		}

		// Check whether the message matches
		if !r.matchMessage.MatchString(m.Trailing()) {
			logger.Trace("Non-Match: Message")
			return false
		}
	}

	// Check whether user has one of the disable rules
	for _, b := range r.DisableOn {
		if badges.Has(b) {
			logger.Tracef("Non-Match: Disable-Badge %s", b)
			return false
		}
	}

	// Check whether user has at least one of the enable rules
	if len(r.EnableOn) > 0 {
		var userHasEnableBadge bool
		for _, b := range r.EnableOn {
			if badges.Has(b) {
				userHasEnableBadge = true
			}
		}
		if !userHasEnableBadge {
			logger.Trace("Non-Match: No enable-badges")
			return false
		}
	}

	// Check on permit
	if r.DisableOnPermit && timerStore.HasPermit(m.User) {
		logger.Trace("Non-Match: Permit")
		return false
	}

	// Check whether rule is in cooldown
	if r.Cooldown != nil && timerStore.Has(r.MatcherID(), *r.Cooldown) {
		logger.Trace("Non-Match: On cooldown")
		return false
	}

	// Nothing objected: Matches!
	return true
}

type ruleAction struct {
	Ban         *string        `yaml:"ban"`
	CounterStep *int64         `yaml:"counter_step"`
	Counter     *string        `yaml:"counter"`
	Respond     *string        `yaml:"respond"`
	Timeout     *time.Duration `yaml:"timeout"`
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

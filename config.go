package main

import (
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/Luzifer/go_helpers/v2/str"
)

var cronParser = cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

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

type autoMessage struct {
	Channel   string `yaml:"channel"`
	Message   string `yaml:"message"`
	UseAction bool   `yaml:"use_action"`

	Cron            string        `yaml:"cron"`
	MessageInterval int64         `yaml:"message_interval"`
	TimeInterval    time.Duration `yaml:"time_interval"`

	disabled              bool
	lastMessageSent       time.Time
	linesSinceLastMessage int64

	lock sync.RWMutex
}

func (a *autoMessage) CanSend() bool {
	if a.disabled || !a.IsValid() {
		return false
	}

	a.lock.RLock()
	defer a.lock.RUnlock()

	switch {
	case a.MessageInterval > a.linesSinceLastMessage:
		// Not enough chatted lines
		return false

	case a.TimeInterval > 0 && a.lastMessageSent.Add(a.TimeInterval).After(time.Now()):
		// Simple timer is not yet expired
		return false

	case a.Cron != "":
		sched, _ := cronParser.Parse(a.Cron)
		if sched.Next(a.lastMessageSent).After(time.Now()) {
			// Cron timer is not yet expired
			return false
		}
	}

	return true
}

func (a *autoMessage) CountMessage(channel string) {
	if strings.TrimLeft(channel, "#") != strings.TrimLeft(a.Channel, "#") {
		return
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	a.linesSinceLastMessage++
}

func (a *autoMessage) ID() string {
	sum := sha1.New()

	fmt.Fprintf(sum, "channel:%q", a.Channel)
	fmt.Fprintf(sum, "message:%q", a.Message)
	fmt.Fprintf(sum, "action:%v", a.UseAction)

	return fmt.Sprintf("sha1:%x", sum.Sum(nil))
}

func (a *autoMessage) IsValid() bool {
	if a.Cron != "" {
		if _, err := cronParser.Parse(a.Cron); err != nil {
			return false
		}
	}

	if a.MessageInterval == 0 && a.TimeInterval == 0 && a.Cron == "" {
		return false
	}

	return true
}

func (a *autoMessage) Send(c *irc.Client) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	msg := a.Message
	if a.UseAction {
		msg = fmt.Sprintf("\001ACTION %s\001", msg)
	}

	if err := c.WriteMessage(&irc.Message{
		Command: "PRIVMSG",
		Params: []string{
			fmt.Sprintf("#%s", strings.TrimLeft(a.Channel, "#")),
			msg,
		},
	}); err != nil {
		return errors.Wrap(err, "sending auto-message")
	}

	a.lastMessageSent = time.Now()
	a.linesSinceLastMessage = 0

	return nil
}

type rule struct {
	Actions []*ruleAction `yaml:"actions"`

	Cooldown        *time.Duration `yaml:"cooldown"`
	SkipCooldownFor []string       `yaml:"skip_cooldown_for"`

	MatchChannels []string `yaml:"match_channels"`
	MatchEvent    *string  `yaml:"match_event"`
	MatchMessage  *string  `yaml:"match_message"`
	MatchUsers    []string `yaml:"match_users"`

	DisableOnMatchMessages []string `yaml:"disable_on_match_messages"`

	DisableOnOffline bool     `yaml:"disable_on_offline"`
	DisableOnPermit  bool     `yaml:"disable_on_permit"`
	DisableOn        []string `yaml:"disable_on"`
	EnableOn         []string `yaml:"enable_on"`

	matchMessage           *regexp.Regexp
	disableOnMatchMessages []*regexp.Regexp
}

func (r rule) MatcherID() string {
	out := sha256.New()

	for _, e := range []*string{
		ptrStr(strings.Join(r.MatchChannels, "|")),
		r.MatchEvent,
		r.MatchMessage,
	} {
		if e != nil {
			fmt.Fprintf(out, *e)
		}
	}

	return fmt.Sprintf("sha256:%x", out.Sum(nil))
}

func (r *rule) Matches(m *irc.Message, event *string) bool {
	var err error

	var (
		badges = ircHandler{}.ParseBadgeLevels(m)
		logger = log.WithFields(log.Fields{
			"msg":  m,
			"rule": r,
		})
	)

	// Check Channel match
	if len(r.MatchChannels) > 0 {
		if len(m.Params) == 0 || !str.StringInSlice(m.Params[0], r.MatchChannels) {
			logger.Trace("Non-Match: Channel")
			return false
		}
	}

	if len(r.MatchUsers) > 0 {
		if !str.StringInSlice(strings.ToLower(m.User), r.MatchUsers) {
			logger.Trace("Non-Match: Users")
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

	if len(r.DisableOnMatchMessages) > 0 {
		// If the regexps were not pre-compiled, do it now
		if len(r.disableOnMatchMessages) != len(r.DisableOnMatchMessages) {
			r.disableOnMatchMessages = nil
			for _, dm := range r.DisableOnMatchMessages {
				dmr, err := regexp.Compile(dm)
				if err != nil {
					logger.WithError(err).Error("Unable to compile expression")
					return false
				}
				r.disableOnMatchMessages = append(r.disableOnMatchMessages, dmr)
			}
		}

		for _, rex := range r.disableOnMatchMessages {
			if rex.MatchString(m.Trailing()) {
				logger.Trace("Non-Match: Disable-On-Message")
				return false
			}
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
	if r.DisableOnPermit && timerStore.HasPermit(m.Params[0], m.User) {
		logger.Trace("Non-Match: Permit")
		return false
	}

	// Check whether rule is in cooldown
	if r.Cooldown != nil && timerStore.InCooldown(r.MatcherID(), *r.Cooldown) {
		var userHasSkipBadge bool
		for _, b := range r.SkipCooldownFor {
			if badges.Has(b) {
				userHasSkipBadge = true
			}
		}
		if !userHasSkipBadge {
			logger.Trace("Non-Match: On cooldown")
			return false
		}
	}

	if r.DisableOnOffline {
		streamLive, err := twitch.HasLiveStream(strings.TrimLeft(m.Params[0], "#"))
		if err != nil {
			logger.WithError(err).Error("Unable to determine live status")
			return false
		}
		if !streamLive {
			logger.Trace("Non-Match: Stream offline")
			return false
		}
	}

	// Nothing objected: Matches!
	return true
}

type ruleAction struct {
	Ban             *string        `json:"ban" yaml:"ban"`
	Command         []string       `json:"command" yaml:"command"`
	CounterSet      *string        `json:"counter_set" yaml:"counter_set"`
	CounterStep     *int64         `json:"counter_step" yaml:"counter_step"`
	Counter         *string        `json:"counter" yaml:"counter"`
	DeleteMessage   *bool          `json:"delete_message" yaml:"delete_message"`
	Respond         *string        `json:"respond" yaml:"respond"`
	RespondFallback *string        `json:"respond_fallback" yaml:"respond_fallback"`
	Timeout         *time.Duration `json:"timeout" yaml:"timeout"`
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

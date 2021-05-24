package main

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/go-irc/irc"
	log "github.com/sirupsen/logrus"
)

type rule struct {
	Actions []*ruleAction `yaml:"actions" hcl:"action,block"`

	Cooldown        *time.Duration `yaml:"cooldown" hcl:"cooldown,optional"`
	SkipCooldownFor []string       `yaml:"skip_cooldown_for" hcl:"skip_cooldown_for,optional"`

	MatchChannels []string `yaml:"match_channels" hcl:"match_channels,optional"`
	MatchEvent    *string  `yaml:"match_event" hcl:"match_event,optional"`
	MatchMessage  *string  `yaml:"match_message" hcl:"match_message,optional"`
	MatchUsers    []string `yaml:"match_users"  hcl:"match_users,optional"`

	DisableOnMatchMessages []string `yaml:"disable_on_match_messages" hcl:"disable_on_match_messages,optional"`

	DisableOnOffline bool     `yaml:"disable_on_offline" hcl:"disable_on_offline,optional"`
	DisableOnPermit  bool     `yaml:"disable_on_permit" hcl:"disable_on_permit,optional"`
	DisableOn        []string `yaml:"disable_on" hcl:"disable_on,optional"`
	EnableOn         []string `yaml:"enable_on" hcl:"enable_on,optional"`

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
	var (
		badges = ircHandler{}.ParseBadgeLevels(m)
		logger = log.WithFields(log.Fields{
			"msg":  m,
			"rule": r,
		})
	)

	for _, matcher := range []func(*log.Entry, *irc.Message, *string, badgeCollection) bool{
		r.allowExecuteChannelWhitelist,
		r.allowExecuteUserWhitelist,
		r.allowExecuteEventWhitelist,
		r.allowExecuteMessageMatcherWhitelist,
		r.allowExecuteMessageMatcherBlacklist,
		r.allowExecuteBadgeBlacklist,
		r.allowExecuteBadgeWhitelist,
		r.allowExecuteDisableOnPermit,
		r.allowExecuteCooldown,
		r.allowExecuteDisableOnOffline,
	} {
		if !matcher(logger, m, event, badges) {
			return false
		}
	}

	// Nothing objected: Matches!
	return true
}

func (r *rule) allowExecuteBadgeBlacklist(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
	for _, b := range r.DisableOn {
		if badges.Has(b) {
			logger.Tracef("Non-Match: Disable-Badge %s", b)
			return false
		}
	}

	return true
}

func (r *rule) allowExecuteBadgeWhitelist(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
	if len(r.EnableOn) == 0 {
		// No match criteria set, does not speak against matching
		return true
	}

	for _, b := range r.EnableOn {
		if badges.Has(b) {
			return true
		}
	}

	return false
}

func (r *rule) allowExecuteChannelWhitelist(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
	if len(r.MatchChannels) == 0 {
		// No match criteria set, does not speak against matching
		return true
	}

	if len(m.Params) == 0 || (!str.StringInSlice(m.Params[0], r.MatchChannels) && !str.StringInSlice(strings.TrimPrefix(m.Params[0], "#"), r.MatchChannels)) {
		logger.Trace("Non-Match: Channel")
		return false
	}

	return true
}

func (r *rule) allowExecuteCooldown(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
	if r.Cooldown == nil {
		// No match criteria set, does not speak against matching
		return true
	}

	if !timerStore.InCooldown(r.MatcherID(), *r.Cooldown) {
		return true
	}

	for _, b := range r.SkipCooldownFor {
		if badges.Has(b) {
			return true
		}
	}

	return false
}

func (r *rule) allowExecuteDisableOnOffline(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
	if !r.DisableOnOffline {
		// No match criteria set, does not speak against matching
		return true
	}

	streamLive, err := twitch.HasLiveStream(strings.TrimLeft(m.Params[0], "#"))
	if err != nil {
		logger.WithError(err).Error("Unable to determine live status")
		return false
	}
	if !streamLive {
		logger.Trace("Non-Match: Stream offline")
		return false
	}

	return true
}

func (r *rule) allowExecuteDisableOnPermit(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
	if r.DisableOnPermit && timerStore.HasPermit(m.Params[0], m.User) {
		logger.Trace("Non-Match: Permit")
		return false
	}

	return true
}

func (r *rule) allowExecuteEventWhitelist(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
	if r.MatchEvent == nil {
		// No match criteria set, does not speak against matching
		return true
	}

	if event == nil || *r.MatchEvent != *event {
		logger.Trace("Non-Match: Event")
		return false
	}

	return true
}

func (r *rule) allowExecuteMessageMatcherBlacklist(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
	if len(r.DisableOnMatchMessages) == 0 {
		// No match criteria set, does not speak against matching
		return true
	}

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

	return true
}

func (r *rule) allowExecuteMessageMatcherWhitelist(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
	if r.MatchMessage == nil {
		// No match criteria set, does not speak against matching
		return true
	}

	var err error

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

	return true
}

func (r *rule) allowExecuteUserWhitelist(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
	if len(r.MatchUsers) == 0 {
		// No match criteria set, does not speak against matching
		return true
	}

	if !str.StringInSlice(strings.ToLower(m.User), r.MatchUsers) {
		logger.Trace("Non-Match: Users")
		return false
	}

	return true
}

type ruleAction struct {
	Ban *string `json:"ban" yaml:"ban" hcl:"ban,optional"`

	Command []string `json:"command" yaml:"command" hcl:"command,optional"`

	CounterSet  *string `json:"counter_set" yaml:"counter_set" hcl:"counter_set,optional"`
	CounterStep *int64  `json:"counter_step" yaml:"counter_step" hcl:"counter_step,optional"`
	Counter     *string `json:"counter" yaml:"counter" hcl:"counter,optional"`

	Delay       time.Duration `json:"delay" yaml:"delay" hcl:"delay,optional"`
	DelayJitter time.Duration `json:"delay_jitter" yaml:"delay_jitter" hcl:"delay_jitter,optional"`

	DeleteMessage *bool `json:"delete_message" yaml:"delete_message" hcl:"delete_message,optional"`

	Respond         *string `json:"respond" yaml:"respond" hcl:"respond,optional"`
	RespondFallback *string `json:"respond_fallback" yaml:"respond_fallback" hcl:"respond_fallback,optional"`

	Timeout *time.Duration `json:"timeout" yaml:"timeout" hcl:"timeout,optional"`
}

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

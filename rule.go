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

type Rule struct {
	Actions []*RuleAction `yaml:"actions"`

	Cooldown        *time.Duration `yaml:"cooldown"`
	ChannelCooldown *time.Duration `yaml:"channel_cooldown"`
	UserCooldown    *time.Duration `yaml:"user_cooldown"`
	SkipCooldownFor []string       `yaml:"skip_cooldown_for"`

	MatchChannels []string `yaml:"match_channels"`
	MatchEvent    *string  `yaml:"match_event"`
	MatchMessage  *string  `yaml:"match_message"`
	MatchUsers    []string `yaml:"match_users" `

	DisableOnMatchMessages []string `yaml:"disable_on_match_messages"`

	Disable           *bool    `yaml:"disable"`
	DisableOnOffline  *bool    `yaml:"disable_on_offline"`
	DisableOnPermit   *bool    `yaml:"disable_on_permit"`
	DisableOnTemplate *string  `yaml:"disable_on_template"`
	DisableOn         []string `yaml:"disable_on"`
	EnableOn          []string `yaml:"enable_on"`

	matchMessage           *regexp.Regexp
	disableOnMatchMessages []*regexp.Regexp
}

func (r Rule) MatcherID() string {
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

func (r *Rule) matches(m *irc.Message, event *string) bool {
	var (
		badges = ircHandler{}.ParseBadgeLevels(m)
		logger = log.WithFields(log.Fields{
			"msg":  m,
			"rule": r,
		})
	)

	for _, matcher := range []func(*log.Entry, *irc.Message, *string, badgeCollection) bool{
		r.allowExecuteDisable,
		r.allowExecuteChannelWhitelist,
		r.allowExecuteUserWhitelist,
		r.allowExecuteEventWhitelist,
		r.allowExecuteMessageMatcherWhitelist,
		r.allowExecuteMessageMatcherBlacklist,
		r.allowExecuteBadgeBlacklist,
		r.allowExecuteBadgeWhitelist,
		r.allowExecuteDisableOnPermit,
		r.allowExecuteRuleCooldown,
		r.allowExecuteChannelCooldown,
		r.allowExecuteUserCooldown,
		r.allowExecuteDisableOnTemplate,
		r.allowExecuteDisableOnOffline,
	} {
		if !matcher(logger, m, event, badges) {
			return false
		}
	}

	// Nothing objected: Matches!
	return true
}

func (r *Rule) setCooldown(m *irc.Message) {
	if r.Cooldown != nil {
		timerStore.AddCooldown(timerTypeCooldown, "", r.MatcherID())
	}

	if r.ChannelCooldown != nil && len(m.Params) > 0 {
		timerStore.AddCooldown(timerTypeCooldown, m.Params[0], r.MatcherID())
	}

	if r.UserCooldown != nil {
		timerStore.AddCooldown(timerTypeCooldown, m.User, r.MatcherID())
	}
}

func (r *Rule) allowExecuteBadgeBlacklist(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
	for _, b := range r.DisableOn {
		if badges.Has(b) {
			logger.Tracef("Non-Match: Disable-Badge %s", b)
			return false
		}
	}

	return true
}

func (r *Rule) allowExecuteBadgeWhitelist(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
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

func (r *Rule) allowExecuteChannelCooldown(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
	if r.ChannelCooldown == nil || len(m.Params) < 1 {
		// No match criteria set, does not speak against matching
		return true
	}

	if !timerStore.InCooldown(timerTypeCooldown, m.Params[0], r.MatcherID(), *r.ChannelCooldown) {
		return true
	}

	for _, b := range r.SkipCooldownFor {
		if badges.Has(b) {
			return true
		}
	}

	return false
}

func (r *Rule) allowExecuteChannelWhitelist(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
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

func (r *Rule) allowExecuteDisable(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
	if r.Disable == nil {
		// No match criteria set, does not speak against matching
		return true
	}

	if *r.Disable {
		logger.Trace("Non-Match: Disable")
		return false
	}

	return true
}

func (r *Rule) allowExecuteDisableOnOffline(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
	if r.DisableOnOffline == nil || !*r.DisableOnOffline {
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

func (r *Rule) allowExecuteDisableOnPermit(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
	if r.DisableOnPermit != nil && *r.DisableOnPermit && timerStore.HasPermit(m.Params[0], m.User) {
		logger.Trace("Non-Match: Permit")
		return false
	}

	return true
}

func (r *Rule) allowExecuteDisableOnTemplate(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
	if r.DisableOnTemplate == nil {
		// No match criteria set, does not speak against matching
		return true
	}

	res, err := formatMessage(*r.DisableOnTemplate, m, r, nil)
	if err != nil {
		logger.WithError(err).Error("Unable to check DisableOnTemplate field")
		// Caused an error, forbid execution
		return false
	}

	if res == "true" {
		logger.Trace("Non-Match: Template")
		return false
	}

	return true
}

func (r *Rule) allowExecuteEventWhitelist(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
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

func (r *Rule) allowExecuteMessageMatcherBlacklist(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
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

func (r *Rule) allowExecuteMessageMatcherWhitelist(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
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

func (r *Rule) allowExecuteRuleCooldown(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
	if r.Cooldown == nil {
		// No match criteria set, does not speak against matching
		return true
	}

	if !timerStore.InCooldown(timerTypeCooldown, "", r.MatcherID(), *r.Cooldown) {
		return true
	}

	for _, b := range r.SkipCooldownFor {
		if badges.Has(b) {
			return true
		}
	}

	return false
}

func (r *Rule) allowExecuteUserCooldown(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
	if r.UserCooldown == nil {
		// No match criteria set, does not speak against matching
		return true
	}

	if !timerStore.InCooldown(timerTypeCooldown, m.User, r.MatcherID(), *r.UserCooldown) {
		return true
	}

	for _, b := range r.SkipCooldownFor {
		if badges.Has(b) {
			return true
		}
	}

	return false
}

func (r *Rule) allowExecuteUserWhitelist(logger *log.Entry, m *irc.Message, event *string, badges badgeCollection) bool {
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

type RuleAction struct{ unmarshal func(interface{}) error }

func (r *RuleAction) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r.unmarshal = unmarshal
	return nil
}

func (r *RuleAction) Unmarshal(v interface{}) error {
	return r.unmarshal(v)
}

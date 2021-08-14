package plugins

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Luzifer/go_helpers/v2/str"
	"github.com/Luzifer/twitch-bot/twitch"
	"github.com/go-irc/irc"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Rule struct {
	UUID string `hash:"-" yaml:"uuid"`

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

	msgFormatter MsgFormatter
	timerStore   TimerStore
	twitchClient *twitch.Client
}

func (r Rule) MatcherID() string {
	if r.UUID != "" {
		return r.UUID
	}

	h, err := hashstructure.Hash(r, hashstructure.FormatV2, nil)
	if err != nil {
		panic(errors.Wrap(err, "hashing automessage"))
	}
	return fmt.Sprintf("hashstructure:%x", h)
}

func (r *Rule) Matches(m *irc.Message, event *string, timerStore TimerStore, msgFormatter MsgFormatter, twitchClient *twitch.Client) bool {
	r.msgFormatter = msgFormatter
	r.timerStore = timerStore
	r.twitchClient = twitchClient

	var (
		badges = twitch.ParseBadgeLevels(m)
		logger = log.WithFields(log.Fields{
			"msg":  m,
			"rule": r,
		})
	)

	for _, matcher := range []func(*log.Entry, *irc.Message, *string, twitch.BadgeCollection) bool{
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

func (r *Rule) GetMatchMessage() *regexp.Regexp {
	var err error

	if r.matchMessage == nil {
		if r.matchMessage, err = regexp.Compile(*r.MatchMessage); err != nil {
			log.WithError(err).Error("Unable to compile expression")
			return nil
		}
	}

	return r.matchMessage
}

func (r *Rule) SetCooldown(timerStore TimerStore, m *irc.Message) {
	if r.Cooldown != nil {
		timerStore.AddCooldown(TimerTypeCooldown, "", r.MatcherID(), time.Now().Add(*r.Cooldown))
	}

	if r.ChannelCooldown != nil && len(m.Params) > 0 {
		timerStore.AddCooldown(TimerTypeCooldown, m.Params[0], r.MatcherID(), time.Now().Add(*r.ChannelCooldown))
	}

	if r.UserCooldown != nil {
		timerStore.AddCooldown(TimerTypeCooldown, m.User, r.MatcherID(), time.Now().Add(*r.UserCooldown))
	}
}

func (r *Rule) allowExecuteBadgeBlacklist(logger *log.Entry, m *irc.Message, event *string, badges twitch.BadgeCollection) bool {
	for _, b := range r.DisableOn {
		if badges.Has(b) {
			logger.Tracef("Non-Match: Disable-Badge %s", b)
			return false
		}
	}

	return true
}

func (r *Rule) allowExecuteBadgeWhitelist(logger *log.Entry, m *irc.Message, event *string, badges twitch.BadgeCollection) bool {
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

func (r *Rule) allowExecuteChannelCooldown(logger *log.Entry, m *irc.Message, event *string, badges twitch.BadgeCollection) bool {
	if r.ChannelCooldown == nil || len(m.Params) < 1 {
		// No match criteria set, does not speak against matching
		return true
	}

	if !r.timerStore.InCooldown(TimerTypeCooldown, m.Params[0], r.MatcherID()) {
		return true
	}

	for _, b := range r.SkipCooldownFor {
		if badges.Has(b) {
			return true
		}
	}

	return false
}

func (r *Rule) allowExecuteChannelWhitelist(logger *log.Entry, m *irc.Message, event *string, badges twitch.BadgeCollection) bool {
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

func (r *Rule) allowExecuteDisable(logger *log.Entry, m *irc.Message, event *string, badges twitch.BadgeCollection) bool {
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

func (r *Rule) allowExecuteDisableOnOffline(logger *log.Entry, m *irc.Message, event *string, badges twitch.BadgeCollection) bool {
	if r.DisableOnOffline == nil || !*r.DisableOnOffline {
		// No match criteria set, does not speak against matching
		return true
	}

	streamLive, err := r.twitchClient.HasLiveStream(strings.TrimLeft(m.Params[0], "#"))
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

func (r *Rule) allowExecuteDisableOnPermit(logger *log.Entry, m *irc.Message, event *string, badges twitch.BadgeCollection) bool {
	if r.DisableOnPermit != nil && *r.DisableOnPermit && r.timerStore.HasPermit(m.Params[0], m.User) {
		logger.Trace("Non-Match: Permit")
		return false
	}

	return true
}

func (r *Rule) allowExecuteDisableOnTemplate(logger *log.Entry, m *irc.Message, event *string, badges twitch.BadgeCollection) bool {
	if r.DisableOnTemplate == nil {
		// No match criteria set, does not speak against matching
		return true
	}

	res, err := r.msgFormatter(*r.DisableOnTemplate, m, r, nil)
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

func (r *Rule) allowExecuteEventWhitelist(logger *log.Entry, m *irc.Message, event *string, badges twitch.BadgeCollection) bool {
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

func (r *Rule) allowExecuteMessageMatcherBlacklist(logger *log.Entry, m *irc.Message, event *string, badges twitch.BadgeCollection) bool {
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

func (r *Rule) allowExecuteMessageMatcherWhitelist(logger *log.Entry, m *irc.Message, event *string, badges twitch.BadgeCollection) bool {
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

func (r *Rule) allowExecuteRuleCooldown(logger *log.Entry, m *irc.Message, event *string, badges twitch.BadgeCollection) bool {
	if r.Cooldown == nil {
		// No match criteria set, does not speak against matching
		return true
	}

	if !r.timerStore.InCooldown(TimerTypeCooldown, "", r.MatcherID()) {
		return true
	}

	for _, b := range r.SkipCooldownFor {
		if badges.Has(b) {
			return true
		}
	}

	return false
}

func (r *Rule) allowExecuteUserCooldown(logger *log.Entry, m *irc.Message, event *string, badges twitch.BadgeCollection) bool {
	if r.UserCooldown == nil {
		// No match criteria set, does not speak against matching
		return true
	}

	if !r.timerStore.InCooldown(TimerTypeCooldown, m.User, r.MatcherID()) {
		return true
	}

	for _, b := range r.SkipCooldownFor {
		if badges.Has(b) {
			return true
		}
	}

	return false
}

func (r *Rule) allowExecuteUserWhitelist(logger *log.Entry, m *irc.Message, event *string, badges twitch.BadgeCollection) bool {
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

package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/irc.v4"
	"gopkg.in/yaml.v3"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
)

const (
	contentTypeJSON = "json"
	contentTypeYAML = "yaml"

	remoteRuleFetchTimeout = 5 * time.Second
)

// ErrStopRuleExecution is a way for actions to terminate execution
// of the current rule gracefully. No actions after this has been
// returned will be executed and no error state will be set
var ErrStopRuleExecution = errors.New("stop rule execution now")

type (
	// Rule represents a rule in the bot configuration
	Rule struct {
		UUID          string  `hash:"-" json:"uuid,omitempty" yaml:"uuid,omitempty"`
		Description   string  `json:"description,omitempty" yaml:"description,omitempty"`
		SubscribeFrom *string `json:"subscribe_from,omitempty" yaml:"subscribe_from,omitempty"`

		Actions []*RuleAction `json:"actions,omitempty" yaml:"actions,omitempty"`

		Cooldown        *time.Duration `json:"cooldown,omitempty" yaml:"cooldown,omitempty"`
		ChannelCooldown *time.Duration `json:"channel_cooldown,omitempty" yaml:"channel_cooldown,omitempty"`
		UserCooldown    *time.Duration `json:"user_cooldown,omitempty" yaml:"user_cooldown,omitempty"`
		SkipCooldownFor []string       `json:"skip_cooldown_for,omitempty" yaml:"skip_cooldown_for,omitempty"`

		MatchChannels []string `json:"match_channels,omitempty" yaml:"match_channels,omitempty"`
		MatchEvent    *string  `json:"match_event,omitempty" yaml:"match_event,omitempty"`
		MatchMessage  *string  `json:"match_message,omitempty" yaml:"match_message,omitempty"`
		MatchUsers    []string `json:"match_users,omitempty" yaml:"match_users,omitempty" `

		DisableOnMatchMessages []string `json:"disable_on_match_messages,omitempty" yaml:"disable_on_match_messages,omitempty"`

		Disable           *bool    `json:"disable,omitempty" yaml:"disable,omitempty"`
		DisableOnOffline  *bool    `json:"disable_on_offline,omitempty" yaml:"disable_on_offline,omitempty"`
		DisableOnPermit   *bool    `json:"disable_on_permit,omitempty" yaml:"disable_on_permit,omitempty"`
		DisableOnTemplate *string  `json:"disable_on_template,omitempty" yaml:"disable_on_template,omitempty"`
		DisableOn         []string `json:"disable_on,omitempty" yaml:"disable_on,omitempty"`
		EnableOn          []string `json:"enable_on,omitempty" yaml:"enable_on,omitempty"`

		//revive:disable-next-line:confusing-naming // only used internally as parsed regexp
		matchMessage *regexp.Regexp
		//revive:disable-next-line:confusing-naming // only used internally as parsed regexp
		disableOnMatchMessages []*regexp.Regexp

		msgFormatter MsgFormatter
		timerStore   TimerStore
		twitchClient *twitch.Client
	}

	// RuleAction represents an action to be executed when running a Rule
	RuleAction struct {
		Type       string                           `json:"type" yaml:"type,omitempty"`
		Attributes *fieldcollection.FieldCollection `json:"attributes" yaml:"attributes,omitempty"`
	}
)

// MatcherID returns the rule UUID or a hash for the rule if no UUID
// is available
func (r Rule) MatcherID() string {
	if r.UUID != "" {
		return r.UUID
	}

	return r.hash()
}

// Matches checks whether the Rule should be executed for the given parameters
func (r *Rule) Matches(m *irc.Message, event *string, timerStore TimerStore, msgFormatter MsgFormatter, twitchClient *twitch.Client, eventData *fieldcollection.FieldCollection) bool {
	r.msgFormatter = msgFormatter
	r.timerStore = timerStore
	r.twitchClient = twitchClient

	var (
		badges = twitch.ParseBadgeLevels(m)
		logger = logrus.WithFields(logrus.Fields{
			"msg":  m,
			"rule": r,
		})
	)

	for _, matcher := range []func(*logrus.Entry, *irc.Message, *string, twitch.BadgeCollection, *fieldcollection.FieldCollection) bool{
		r.allowExecuteDisable,
		r.allowExecuteChannelWhitelist,
		r.allowExecuteUserWhitelist,
		r.allowExecuteEventMatch,
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
		if !matcher(logger, m, event, badges, eventData) {
			return false
		}
	}

	// Nothing objected: Matches!
	return true
}

// GetMatchMessage returns the cached Regexp if available or compiles
// the given match string into a Regexp
func (r *Rule) GetMatchMessage() *regexp.Regexp {
	var err error

	if r.matchMessage == nil {
		if r.matchMessage, err = regexp.Compile(*r.MatchMessage); err != nil {
			logrus.WithError(err).Error("Unable to compile expression")
			return nil
		}
	}

	return r.matchMessage
}

// SetCooldown uses the given TimerStore to set the cooldowns for the
// Rule after execution
func (r *Rule) SetCooldown(timerStore TimerStore, m *irc.Message, evtData *fieldcollection.FieldCollection) {
	var err error

	if r.Cooldown != nil {
		if err = timerStore.AddCooldown(TimerTypeCooldown, "", r.MatcherID(), time.Now().Add(*r.Cooldown)); err != nil {
			logrus.WithError(err).Error("setting general rule cooldown")
		}
	}

	if r.ChannelCooldown != nil && DeriveChannel(m, evtData) != "" {
		if err = timerStore.AddCooldown(TimerTypeCooldown, DeriveChannel(m, evtData), r.MatcherID(), time.Now().Add(*r.ChannelCooldown)); err != nil {
			logrus.WithError(err).Error("setting channel rule cooldown")
		}
	}

	if r.UserCooldown != nil && DeriveUser(m, evtData) != "" {
		if err = timerStore.AddCooldown(TimerTypeCooldown, DeriveUser(m, evtData), r.MatcherID(), time.Now().Add(*r.UserCooldown)); err != nil {
			logrus.WithError(err).Error("setting user rule cooldown")
		}
	}
}

// UpdateFromSubscription fetches the remote Rule source if one is
// defined and updates the rule with its content
func (r *Rule) UpdateFromSubscription(ctx context.Context) (bool, error) {
	if r.SubscribeFrom == nil || len(*r.SubscribeFrom) == 0 {
		return false, nil
	}

	prevHash := r.hash()

	remoteURL, err := url.Parse(*r.SubscribeFrom)
	if err != nil {
		return false, errors.Wrap(err, "parsing remote subscription url")
	}

	reqCtx, cancel := context.WithTimeout(ctx, remoteRuleFetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, remoteURL.String(), nil)
	if err != nil {
		return false, errors.Wrap(err, "assembling request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, errors.Wrap(err, "executing request")
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Error("closing request body (leaked fd)")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return false, errors.Errorf("unxpected HTTP status %d", resp.StatusCode)
	}

	inputType, err := r.fileTypeFromRequest(remoteURL, resp)
	if err != nil {
		return false, errors.Wrap(err, "detecting content type")
	}

	var newRule Rule
	switch inputType {
	case contentTypeJSON:
		err = json.NewDecoder(resp.Body).Decode(&newRule)

	case contentTypeYAML:
		err = yaml.NewDecoder(resp.Body).Decode(&newRule)

	default:
		return false, errors.New("unexpected format")
	}

	if err != nil {
		return false, errors.Wrap(err, "decoding remote rule")
	}

	if newRule.hash() == prevHash {
		// No update, exit now
		return false, nil
	}

	*r = newRule

	return true, nil
}

// Validate executes some basic checks on the validity of the Rule
func (r Rule) Validate(tplValidate TemplateValidatorFunc) error {
	if r.MatchMessage != nil {
		if _, err := regexp.Compile(*r.MatchMessage); err != nil {
			return errors.Wrap(err, "compiling match_message field regex")
		}
	}

	if r.DisableOnTemplate != nil {
		if err := tplValidate(*r.DisableOnTemplate); err != nil {
			return errors.Wrap(err, "parsing disable_on_template template")
		}
	}

	return nil
}

func (r *Rule) allowExecuteBadgeBlacklist(logger *logrus.Entry, _ *irc.Message, _ *string, badges twitch.BadgeCollection, _ *fieldcollection.FieldCollection) bool {
	for _, b := range r.DisableOn {
		if badges.Has(b) {
			logger.Tracef("Non-Match: Disable-Badge %s", b)
			return false
		}
	}

	return true
}

func (r *Rule) allowExecuteBadgeWhitelist(_ *logrus.Entry, _ *irc.Message, _ *string, badges twitch.BadgeCollection, _ *fieldcollection.FieldCollection) bool {
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

func (r *Rule) allowExecuteChannelCooldown(logger *logrus.Entry, m *irc.Message, _ *string, badges twitch.BadgeCollection, evtData *fieldcollection.FieldCollection) bool {
	if r.ChannelCooldown == nil || DeriveChannel(m, evtData) == "" {
		// No match criteria set, does not speak against matching
		return true
	}

	inCooldown, err := r.timerStore.InCooldown(TimerTypeCooldown, DeriveChannel(m, evtData), r.MatcherID())
	if err != nil {
		logger.WithError(err).Error("checking channel cooldown")
		return false
	}

	if !inCooldown {
		return true
	}

	for _, b := range r.SkipCooldownFor {
		if badges.Has(b) {
			return true
		}
	}

	return false
}

func (r *Rule) allowExecuteChannelWhitelist(logger *logrus.Entry, m *irc.Message, _ *string, _ twitch.BadgeCollection, evtData *fieldcollection.FieldCollection) bool {
	if len(r.MatchChannels) == 0 {
		// No match criteria set, does not speak against matching
		return true
	}

	if DeriveChannel(m, evtData) == "" || (!slices.Contains(r.MatchChannels, DeriveChannel(m, evtData)) && !slices.Contains(r.MatchChannels, strings.TrimPrefix(DeriveChannel(m, evtData), "#"))) {
		logger.Trace("Non-Match: Channel")
		return false
	}

	return true
}

func (r *Rule) allowExecuteDisable(logger *logrus.Entry, _ *irc.Message, _ *string, _ twitch.BadgeCollection, _ *fieldcollection.FieldCollection) bool {
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

func (r *Rule) allowExecuteDisableOnOffline(logger *logrus.Entry, m *irc.Message, _ *string, _ twitch.BadgeCollection, evtData *fieldcollection.FieldCollection) bool {
	if r.DisableOnOffline == nil || !*r.DisableOnOffline || DeriveChannel(m, evtData) == "" {
		// No match criteria set, does not speak against matching
		return true
	}

	streamLive, err := r.twitchClient.HasLiveStream(context.Background(), strings.TrimLeft(DeriveChannel(m, evtData), "#"))
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

func (r *Rule) allowExecuteDisableOnPermit(logger *logrus.Entry, m *irc.Message, _ *string, _ twitch.BadgeCollection, evtData *fieldcollection.FieldCollection) bool {
	hasPermit, err := r.timerStore.HasPermit(DeriveChannel(m, evtData), DeriveUser(m, evtData))
	if err != nil {
		logger.WithError(err).Error("checking permit")
		return false
	}

	if r.DisableOnPermit != nil && *r.DisableOnPermit && DeriveChannel(m, evtData) != "" && hasPermit {
		logger.Trace("Non-Match: Permit")
		return false
	}

	return true
}

func (r *Rule) allowExecuteDisableOnTemplate(logger *logrus.Entry, m *irc.Message, _ *string, _ twitch.BadgeCollection, evtData *fieldcollection.FieldCollection) bool {
	if r.DisableOnTemplate == nil || *r.DisableOnTemplate == "" {
		// No match criteria set, does not speak against matching
		return true
	}

	res, err := r.msgFormatter(*r.DisableOnTemplate, m, r, evtData)
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

func (r *Rule) allowExecuteEventMatch(logger *logrus.Entry, _ *irc.Message, event *string, _ twitch.BadgeCollection, _ *fieldcollection.FieldCollection) bool {
	// The user defines either no event to match or they define an
	// event to match. We now need to ensure this match is valid for
	// the current execution:
	//
	// - If the user gave no event we MUST NOT have an event defined
	// - If the user gave an event we MUST have the same event defined
	//
	// To aid fighting spam we do define some excemption from these
	// rules:
	//
	// - Bits are sent using IRC messages and might contains spam
	//   therefore we additionally match them through a message
	//   matcher
	// - Resubs are also IRC messages and might be abused to spam
	//   though this is quite unlikely. Even though it's unlikely
	//   we also allow a match for message matchers to aid mods
	//
	// As all set events are always pointer to non-empty strings we
	// assume an empty string in case either is not set and then
	// compare the string contents.

	var mE, gE string

	if r.MatchEvent != nil {
		mE = *r.MatchEvent
	}

	if event != nil {
		gE = *event
	}

	if mE == gE {
		// Event does exactly match
		return true
	}

	if mE == "" && slices.Contains([]string{"bits", "resub"}, gE) {
		// Additional message matchers - see explanation above
		return true
	}

	logger.Trace("Non-Match: Event")
	return false
}

func (r *Rule) allowExecuteMessageMatcherBlacklist(logger *logrus.Entry, m *irc.Message, _ *string, _ twitch.BadgeCollection, _ *fieldcollection.FieldCollection) bool {
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
		if m != nil && rex.MatchString(m.Trailing()) {
			logger.Trace("Non-Match: Disable-On-Message")
			return false
		}
	}

	return true
}

func (r *Rule) allowExecuteMessageMatcherWhitelist(logger *logrus.Entry, m *irc.Message, _ *string, _ twitch.BadgeCollection, _ *fieldcollection.FieldCollection) bool {
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
	if m == nil || !r.matchMessage.MatchString(m.Trailing()) {
		logger.Trace("Non-Match: Message")
		return false
	}

	return true
}

func (r *Rule) allowExecuteRuleCooldown(logger *logrus.Entry, _ *irc.Message, _ *string, badges twitch.BadgeCollection, _ *fieldcollection.FieldCollection) bool {
	if r.Cooldown == nil {
		// No match criteria set, does not speak against matching
		return true
	}

	inCooldown, err := r.timerStore.InCooldown(TimerTypeCooldown, "", r.MatcherID())
	if err != nil {
		logger.WithError(err).Error("checking rule cooldown")
		return false
	}

	if !inCooldown {
		return true
	}

	for _, b := range r.SkipCooldownFor {
		if badges.Has(b) {
			return true
		}
	}

	return false
}

func (r *Rule) allowExecuteUserCooldown(logger *logrus.Entry, m *irc.Message, _ *string, badges twitch.BadgeCollection, evtData *fieldcollection.FieldCollection) bool {
	if r.UserCooldown == nil {
		// No match criteria set, does not speak against matching
		return true
	}

	inCooldown, err := r.timerStore.InCooldown(TimerTypeCooldown, DeriveUser(m, evtData), r.MatcherID())
	if err != nil {
		logger.WithError(err).Error("checking user cooldown")
		return false
	}

	if DeriveUser(m, evtData) == "" || !inCooldown {
		return true
	}

	for _, b := range r.SkipCooldownFor {
		if badges.Has(b) {
			return true
		}
	}

	return false
}

func (r *Rule) allowExecuteUserWhitelist(logger *logrus.Entry, m *irc.Message, _ *string, _ twitch.BadgeCollection, evtData *fieldcollection.FieldCollection) bool {
	if len(r.MatchUsers) == 0 {
		// No match criteria set, does not speak against matching
		return true
	}

	if DeriveUser(m, evtData) == "" || !slices.Contains(r.MatchUsers, strings.ToLower(DeriveUser(m, evtData))) {
		logger.Trace("Non-Match: Users")
		return false
	}

	return true
}

func (Rule) fileTypeFromRequest(remoteURL *url.URL, resp *http.Response) (string, error) {
	switch path.Ext(remoteURL.Path) {
	case ".json":
		return contentTypeJSON, nil

	case ".yaml", ".yml":
		return contentTypeYAML, nil
	}

	switch strings.Split(resp.Header.Get("Content-Type"), ";")[0] {
	case "application/json":
		return contentTypeJSON, nil

	case "application/yaml", "application/x-yaml", "text/x-yaml":
		return contentTypeYAML, nil
	}

	return "", errors.New("no valid file type detected")
}

func (r Rule) hash() string {
	h, err := hashstructure.Hash(r, hashstructure.FormatV2, nil)
	if err != nil {
		panic(errors.Wrap(err, "hashing rule"))
	}
	return fmt.Sprintf("hashstructure:%x", h)
}

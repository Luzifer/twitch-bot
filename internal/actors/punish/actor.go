package punish

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

const (
	actorNamePunish      = "punish"
	actorNameResetPunish = "reset-punish"
	moduleUUID           = "44ab4646-ce50-4e16-9353-c1f0eb68962b"
)

var (
	formatMessage      plugins.MsgFormatter
	ptrDefaultCooldown = func(v time.Duration) *time.Duration { return &v }(168 * time.Hour)
	ptrStringEmpty     = func(v string) *string { return &v }("")
	store              plugins.StorageManager
	storedObject       = newStorage()
)

func Register(args plugins.RegistrationArguments) error {
	formatMessage = args.FormatMessage
	store = args.GetStorageManager()

	args.RegisterActor(actorNamePunish, func() plugins.Actor { return &actorPunish{} })
	args.RegisterActor(actorNameResetPunish, func() plugins.Actor { return &actorResetPunish{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Apply increasing punishments to user",
		Name:        "Punish User",
		Type:        actorNamePunish,

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "168h",
				Description:     "When to lower the punishment level after the last punishment",
				Key:             "cooldown",
				Name:            "Cooldown",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeDuration,
			},
			{
				Default:         "",
				Description:     "Actions for each punishment level (ban, delete, duration-value i.e. 1m)",
				Key:             "levels",
				Name:            "Levels",
				Optional:        false,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeStringSlice,
			},
			{
				Default:         "",
				Description:     "Reason why the user was banned",
				Key:             "reason",
				Name:            "Reason",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "",
				Description:     "User to apply the action to",
				Key:             "user",
				Name:            "User",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "",
				Description:     "Unique identifier for this punishment to differentiate between punishments in the same channel",
				Key:             "uuid",
				Name:            "UUID",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Reset punishment level for user",
		Name:        "Reset User Punishment",
		Type:        actorNameResetPunish,

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "User to reset the level for",
				Key:             "user",
				Name:            "User",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "",
				Description:     "Unique identifier for this punishment to differentiate between punishments in the same channel",
				Key:             "uuid",
				Name:            "UUID",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	return errors.Wrap(
		store.GetModuleStore(moduleUUID, storedObject),
		"loading module storage",
	)
}

type (
	actorPunish      struct{}
	actorResetPunish struct{}

	levelConfig struct {
		LastLevel int           `json:"last_level"`
		Executed  time.Time     `json:"executed"`
		Cooldown  time.Duration `json:"cooldown"`
	}

	storage struct {
		ActiveLevels map[string]*levelConfig `json:"active_levels"`

		lock sync.RWMutex
	}
)

// Punish

func (a actorPunish) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData plugins.FieldCollection, attrs plugins.FieldCollection) (preventCooldown bool, err error) {
	var (
		cooldown = attrs.MustDuration("cooldown", ptrDefaultCooldown)
		reason   = attrs.MustString("reason", ptrStringEmpty)
		user     = attrs.MustString("user", nil)
		uuid     = attrs.MustString("uuid", ptrStringEmpty)
	)

	levels, err := attrs.StringSlice("levels")
	if err != nil {
		return false, errors.Wrap(err, "getting level config")
	}

	if user, err = formatMessage(user, m, r, eventData); err != nil {
		return false, errors.Wrap(err, "preparing user")
	}

	lvl := storedObject.GetPunishment(plugins.DeriveChannel(m, eventData), user, uuid)
	nLvl := int(math.Min(float64(len(levels)-1), float64(lvl.LastLevel+1)))

	var cmd []string

	switch lt := levels[nLvl]; lt {
	case "ban":
		cmd = []string{"/ban", strings.TrimLeft(user, "@")}
		if reason != "" {
			cmd = append(cmd, reason)
		}

	case "delete":
		msgID, ok := m.Tags.GetTag("id")
		if !ok || msgID == "" {
			return false, errors.New("found no mesage id")
		}

		cmd = []string{"/delete", msgID}

	default:
		to, err := time.ParseDuration(lt)
		if err != nil {
			return false, errors.Wrap(err, "parsing punishment level")
		}

		cmd = []string{"/timeout", strings.TrimLeft(user, "@"), strconv.FormatInt(int64(to/time.Second), 10)}
		if reason != "" {
			cmd = append(cmd, reason)
		}
	}

	if err := c.WriteMessage(&irc.Message{
		Command: "PRIVMSG",
		Params: []string{
			plugins.DeriveChannel(m, eventData),
			strings.Join(cmd, " "),
		},
	}); err != nil {
		return false, errors.Wrap(err, "sending command")
	}

	lvl.Cooldown = cooldown
	lvl.Executed = time.Now()
	lvl.LastLevel = nLvl

	return false, errors.Wrap(
		store.SetModuleStore(moduleUUID, storedObject),
		"storing punishment level",
	)
}

func (a actorPunish) IsAsync() bool { return false }
func (a actorPunish) Name() string  { return actorNamePunish }

func (a actorPunish) Validate(attrs plugins.FieldCollection) (err error) {
	if v, err := attrs.String("user"); err != nil || v == "" {
		return errors.New("user must be non-empty string")
	}

	if v, err := attrs.StringSlice("levels"); err != nil || len(v) == 0 {
		return errors.New("levels must be slice of strings with length > 0")
	}

	return nil
}

// Reset

func (a actorResetPunish) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData plugins.FieldCollection, attrs plugins.FieldCollection) (preventCooldown bool, err error) {
	var (
		user = attrs.MustString("user", nil)
		uuid = attrs.MustString("uuid", ptrStringEmpty)
	)

	if user, err = formatMessage(user, m, r, eventData); err != nil {
		return false, errors.Wrap(err, "preparing user")
	}

	storedObject.ResetLevel(plugins.DeriveChannel(m, eventData), user, uuid)

	return false, errors.Wrap(
		store.SetModuleStore(moduleUUID, storedObject),
		"resetting punishment level",
	)
}

func (a actorResetPunish) IsAsync() bool { return false }
func (a actorResetPunish) Name() string  { return actorNameResetPunish }

func (a actorResetPunish) Validate(attrs plugins.FieldCollection) (err error) {
	if v, err := attrs.String("user"); err != nil || v == "" {
		return errors.New("user must be non-empty string")
	}

	return nil
}

// Storage

func newStorage() *storage {
	return &storage{}
}

func (s *storage) GetPunishment(channel, user, uuid string) *levelConfig {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Ensure old state is cleared
	s.calculateCooldowns()

	var (
		id  = s.getCacheKey(channel, user, uuid)
		lvl = s.ActiveLevels[id]
	)

	if lvl == nil {
		// Initalize a non-triggered state
		lvl = &levelConfig{LastLevel: -1}
		s.ActiveLevels[id] = lvl
	}

	return lvl
}

func (s *storage) ResetLevel(channel, user, uuid string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.ActiveLevels, s.getCacheKey(channel, user, uuid))
}

func (s *storage) getCacheKey(channel, user, uuid string) string {
	return strings.Join([]string{channel, user, uuid}, "::")
}

func (s *storage) calculateCooldowns() {
	// This MUST NOT be locked, the lock MUST be set by calling method

	var clear []string

	for id, lvl := range s.ActiveLevels {
		for {
			cooldownTime := lvl.Executed.Add(-lvl.Cooldown)
			if cooldownTime.After(time.Now()) {
				break
			}

			lvl.Executed = cooldownTime
			lvl.LastLevel--
		}

		// Level 0 is the first punishment level, so only remove if it drops below 0
		if lvl.LastLevel < 0 {
			clear = append(clear, id)
		}
	}

	for _, id := range clear {
		delete(s.ActiveLevels, id)
	}
}

// Implement JSON marshaller interfaces
func (s *storage) MarshalJSON() ([]byte, error) {
	s.calculateCooldowns()
	return json.Marshal(s)
}

func (s *storage) UnmarshalJSON(data []byte) error { return json.Unmarshal(data, s) }

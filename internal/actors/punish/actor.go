// Package punish contains an actor to punish behaviour in a channel
// with rising punishments
package punish

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"gopkg.in/irc.v4"
	"gorm.io/gorm"

	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/database"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const (
	actorNamePunish      = "punish"
	actorNameResetPunish = "reset-punish"

	oneWeek = 168 * time.Hour
)

type (
	actorPunish      struct{}
	actorResetPunish struct{}

	levelConfig struct {
		LastLevel int           `json:"last_level"`
		Executed  time.Time     `json:"executed"`
		Cooldown  time.Duration `json:"cooldown"`
	}
)

var (
	botTwitchClient func() *twitch.Client
	db              database.Connector
	formatMessage   plugins.MsgFormatter
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	db = args.GetDatabaseConnector()
	if err := db.DB().AutoMigrate(&punishLevel{}); err != nil {
		return fmt.Errorf("applying schema migration: %w", err)
	}

	args.RegisterCopyDatabaseFunc("punish", func(src, target *gorm.DB) error {
		return database.CopyObjects(src, target, &punishLevel{})
	})

	botTwitchClient = args.GetTwitchClient
	formatMessage = args.FormatMessage

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
				Description:     "Reason why the user was banned / timeouted",
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

	return nil
}

// Punish

func (actorPunish) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	var (
		cooldown = attrs.MustDuration("cooldown", helpers.Ptr(oneWeek))
		reason   = attrs.MustString("reason", helpers.Ptr(""))
		user     = attrs.MustString("user", nil)
		uuid     = attrs.MustString("uuid", helpers.Ptr(""))
	)

	levels, err := attrs.StringSlice("levels")
	if err != nil {
		return false, fmt.Errorf("getting level config: %w", err)
	}

	if user, err = formatMessage(user, m, r, eventData); err != nil {
		return false, fmt.Errorf("preparing user: %w", err)
	}

	lvl, err := getPunishment(db, plugins.DeriveChannel(m, eventData), user, uuid)
	if err != nil {
		return false, fmt.Errorf("getting stored punishment: %w", err)
	}
	nLvl := int(math.Min(float64(len(levels)-1), float64(lvl.LastLevel+1)))

	switch lt := levels[nLvl]; lt {
	case "ban":
		if err = botTwitchClient().BanUser(
			context.Background(),
			plugins.DeriveChannel(m, eventData),
			strings.TrimLeft(user, "@"),
			0,
			reason,
		); err != nil {
			return false, fmt.Errorf("executing user ban: %w", err)
		}

	case "delete":
		msgID, ok := m.Tags["id"]
		if !ok || msgID == "" {
			return false, errors.New("found no mesage id")
		}

		if err = botTwitchClient().DeleteMessage(
			context.Background(),
			plugins.DeriveChannel(m, eventData),
			msgID,
		); err != nil {
			return false, fmt.Errorf("deleting message: %w", err)
		}

	default:
		to, err := time.ParseDuration(lt)
		if err != nil {
			return false, fmt.Errorf("parsing punishment level: %w", err)
		}

		if err = botTwitchClient().BanUser(
			context.Background(),
			plugins.DeriveChannel(m, eventData),
			strings.TrimLeft(user, "@"),
			to,
			reason,
		); err != nil {
			return false, fmt.Errorf("executing user ban: %w", err)
		}
	}

	lvl.Cooldown = cooldown
	lvl.Executed = time.Now().UTC()
	lvl.LastLevel = nLvl

	if err = setPunishment(db, plugins.DeriveChannel(m, eventData), user, uuid, lvl); err != nil {
		return false, fmt.Errorf("storing punishment level: %w", err)
	}

	return false, nil
}

func (actorPunish) IsAsync() bool { return false }
func (actorPunish) Name() string  { return actorNamePunish }

func (actorPunish) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "levels", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeStringSlice}),
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "user", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "cooldown", Type: fieldcollection.SchemaFieldTypeDuration}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "reason", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "uuid", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.MustHaveNoUnknowFields,
		helpers.SchemaValidateTemplateField(tplValidator, "user"),
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	return nil
}

// Reset

func (actorResetPunish) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	var (
		user = attrs.MustString("user", nil)
		uuid = attrs.MustString("uuid", helpers.Ptr(""))
	)

	if user, err = formatMessage(user, m, r, eventData); err != nil {
		return false, fmt.Errorf("preparing user: %w", err)
	}

	if err = deletePunishment(db, plugins.DeriveChannel(m, eventData), user, uuid); err != nil {
		return false, fmt.Errorf("resetting punishment level: %w", err)
	}

	return false, nil
}

func (actorResetPunish) IsAsync() bool { return false }
func (actorResetPunish) Name() string  { return actorNameResetPunish }

func (actorResetPunish) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "user", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "uuid", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.MustHaveNoUnknowFields,
		helpers.SchemaValidateTemplateField(tplValidator, "user"),
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	return nil
}

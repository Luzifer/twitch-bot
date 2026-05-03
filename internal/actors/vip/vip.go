// Package vip contains actors to modify VIPs of a channel
package vip

import (
	"context"
	"fmt"
	"strings"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

type (
	actor      struct{}
	unvipActor struct{ actor }
	vipActor   struct{ actor }
)

var (
	formatMessage plugins.MsgFormatter
	permCheckFn   plugins.ChannelPermissionCheckFunc
	tcGetter      func(string) (*twitch.Client, error)
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	formatMessage = args.FormatMessage
	permCheckFn = args.HasPermissionForChannel
	tcGetter = args.GetTwitchClientForChannel

	args.RegisterActor("vip", func() plugins.Actor { return &vipActor{} })
	args.RegisterActor("unvip", func() plugins.Actor { return &unvipActor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Add VIP for the given channel",
		Name:        "Add VIP",
		Type:        "vip",

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Channel to add the VIP to",
				Key:             "channel",
				Name:            "Channel",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "",
				Description:     "User to add as VIP",
				Key:             "user",
				Name:            "User",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Remove VIP for the given channel",
		Name:        "Remove VIP",
		Type:        "unvip",

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Channel to remove the VIP from",
				Key:             "channel",
				Name:            "Channel",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "",
				Description:     "User to remove as VIP",
				Key:             "user",
				Name:            "User",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	args.RegisterMessageModFunc("/vip", handleAddVIP)
	args.RegisterMessageModFunc("/unvip", handleRemoveVIP)

	return nil
}

func (actor) IsAsync() bool { return false }
func (actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "channel", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "user", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.MustHaveNoUnknowFields,
		helpers.SchemaValidateTemplateField(tplValidator, "channel", "user"),
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	return nil
}

func (actor) getParams(m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (channel, user string, err error) {
	if channel, err = formatMessage(attrs.MustString("channel", nil), m, r, eventData); err != nil {
		return "", "", fmt.Errorf("parsing channel: %w", err)
	}

	if user, err = formatMessage(attrs.MustString("user", nil), m, r, eventData); err != nil {
		return "", "", fmt.Errorf("parsing user: %w", err)
	}

	return strings.TrimLeft(channel, "#"), user, nil
}

func (u unvipActor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	channel, user, err := u.getParams(m, r, eventData, attrs)
	if err != nil {
		return false, fmt.Errorf("getting parameters: %w", err)
	}

	if err = executeModVIP(channel, func(tc *twitch.Client) error {
		return tc.RemoveChannelVIP(context.Background(), channel, user)
	}); err != nil {
		return false, fmt.Errorf("removing VIP: %w", err)
	}

	return false, nil
}

func (unvipActor) Name() string { return "unvip" }

func (v vipActor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	channel, user, err := v.getParams(m, r, eventData, attrs)
	if err != nil {
		return false, fmt.Errorf("getting parameters: %w", err)
	}

	if err = executeModVIP(channel, func(tc *twitch.Client) error {
		return tc.AddChannelVIP(context.Background(), channel, user)
	}); err != nil {
		return false, fmt.Errorf("adding VIP: %w", err)
	}

	return false, nil
}

func (vipActor) Name() string { return "vip" }

// Generic helper

func executeModVIP(channel string, modFn func(tc *twitch.Client) error) error {
	ok, err := permCheckFn(channel, twitch.ScopeChannelManageVIPS)
	if err != nil {
		return fmt.Errorf("checking for channel permissions: %w", err)
	}

	if !ok {
		return fmt.Errorf("channel %q is missing permission %s", channel, twitch.ScopeChannelManageVIPS)
	}

	tc, err := tcGetter(channel)
	if err != nil {
		return fmt.Errorf("getting channel twitch-client: %w", err)
	}

	return modFn(tc)
}

// Chat-Commands

func handleAddVIP(m *irc.Message) error {
	return handleModVIP(m, func(tc *twitch.Client, channel, user string) error {
		if err := tc.AddChannelVIP(context.Background(), channel, user); err != nil {
			return fmt.Errorf("adding VIP: %w", err)
		}

		return nil
	})
}

func handleModVIP(m *irc.Message, modFn func(tc *twitch.Client, channel, user string) error) error {
	channel := strings.TrimLeft(plugins.DeriveChannel(m, nil), "#")

	parts := strings.Split(m.Trailing(), " ")
	if len(parts) != 2 { //nolint:mnd // Just a count, makes no sense as a constant
		return fmt.Errorf("wrong command usage, must consist of 2 words")
	}

	return executeModVIP(channel, func(tc *twitch.Client) error { return modFn(tc, channel, parts[1]) })
}

func handleRemoveVIP(m *irc.Message) error {
	return handleModVIP(m, func(tc *twitch.Client, channel, user string) error {
		if err := tc.RemoveChannelVIP(context.Background(), channel, user); err != nil {
			return fmt.Errorf("removing VIP: %w", err)
		}

		return nil
	})
}

package vip

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

var (
	formatMessage plugins.MsgFormatter
	permCheckFn   plugins.ChannelPermissionCheckFunc
	tcGetter      func(string) (*twitch.Client, error)

	ptrStringEmpty = func(s string) *string { return &s }("")
)

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

// Actor

type (
	actor      struct{}
	unvipActor struct{ actor }
	vipActor   struct{ actor }
)

func (actor) IsAsync() bool { return false }
func (a actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *plugins.FieldCollection) (err error) {
	for _, field := range []string{"channel", "user"} {
		if v, err := attrs.String(field); err != nil || v == "" {
			return errors.Errorf("%s must be non-empty string", field)
		}

		if err = tplValidator(attrs.MustString(field, ptrStringEmpty)); err != nil {
			return errors.Wrapf(err, "validating %s template", field)
		}
	}

	return nil
}

func (a actor) getParams(m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (channel, user string, err error) {
	if channel, err = formatMessage(attrs.MustString("channel", nil), m, r, eventData); err != nil {
		return "", "", errors.Wrap(err, "parsing channel")
	}

	if user, err = formatMessage(attrs.MustString("user", nil), m, r, eventData); err != nil {
		return "", "", errors.Wrap(err, "parsing user")
	}

	return strings.TrimLeft(channel, "#"), user, nil
}

func (u unvipActor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	channel, user, err := u.getParams(m, r, eventData, attrs)
	if err != nil {
		return false, errors.Wrap(err, "getting parameters")
	}

	return false, errors.Wrap(
		executeModVIP(channel, func(tc *twitch.Client) error { return tc.RemoveChannelVIP(context.Background(), channel, user) }),
		"removing VIP",
	)
}

func (unvipActor) Name() string { return "unvip" }

func (v vipActor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	channel, user, err := v.getParams(m, r, eventData, attrs)
	if err != nil {
		return false, errors.Wrap(err, "getting parameters")
	}

	return false, errors.Wrap(
		executeModVIP(channel, func(tc *twitch.Client) error { return tc.AddChannelVIP(context.Background(), channel, user) }),
		"adding VIP",
	)
}

func (vipActor) Name() string { return "vip" }

// Generic helper

func executeModVIP(channel string, modFn func(tc *twitch.Client) error) error {
	ok, err := permCheckFn(channel, twitch.ScopeChannelManageVIPS)
	if err != nil {
		return errors.Wrap(err, "checking for channel permissions")
	}

	if !ok {
		return errors.Errorf("channel %q is missing permission %s", channel, twitch.ScopeChannelManageVIPS)
	}

	tc, err := tcGetter(channel)
	if err != nil {
		return errors.Wrap(err, "getting channel twitch-client")
	}

	return modFn(tc)
}

// Chat-Commands

func handleAddVIP(m *irc.Message) error {
	return handleModVIP(m, func(tc *twitch.Client, channel, user string) error {
		return errors.Wrap(tc.AddChannelVIP(context.Background(), channel, user), "adding VIP")
	})
}

func handleModVIP(m *irc.Message, modFn func(tc *twitch.Client, channel, user string) error) error {
	channel := strings.TrimLeft(plugins.DeriveChannel(m, nil), "#")

	parts := strings.Split(m.Trailing(), " ")
	if len(parts) != 2 { //nolint:gomnd // Just a count, makes no sense as a constant
		return errors.Errorf("wrong command usage, must consist of 2 words")
	}

	return executeModVIP(channel, func(tc *twitch.Client) error { return modFn(tc, channel, parts[1]) })
}

func handleRemoveVIP(m *irc.Message) error {
	return handleModVIP(m, func(tc *twitch.Client, channel, user string) error {
		return errors.Wrap(tc.RemoveChannelVIP(context.Background(), channel, user), "removing VIP")
	})
}

// Package clip contains an actor to create clips on behalf of a
// channels owner
package clip

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "clip"

var (
	formatMessage plugins.MsgFormatter
	hasPerm       plugins.ChannelPermissionCheckFunc
	tcGetter      func(string) (*twitch.Client, error)
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	formatMessage = args.FormatMessage
	hasPerm = args.HasPermissionForChannel
	tcGetter = args.GetTwitchClientForChannel

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Triggers the creation of a Clip from the given channel owned by the creator (subsequent actions can use variables `create_clip_slug` and `create_clip_edit_url`)",
		Name:        "Create Clip",
		Type:        actorName,

		Fields: []plugins.ActionDocumentationField{
			{
				Description:     "Channel to create the clip from, defaults to the channel of the event / message",
				Key:             "channel",
				Name:            "Channel",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Description:     fmt.Sprintf("User which should trigger and therefore own the clip (must have given %s permission to the bot in extended permissions!), defaults to the value of `channel`", twitch.ScopeClipsEdit),
				Key:             "creator",
				Name:            "Creator",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "false",
				Description:     "Whether to add an artificial delay before creating the clip",
				Key:             "add_delay",
				Name:            "Add Delay",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeBool,
			},
		},
	})

	return nil
}

type actor struct{}

func (actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	channel := plugins.DeriveChannel(m, eventData)
	if channel, err = formatMessage(attrs.MustString("channel", &channel), m, r, eventData); err != nil {
		return false, errors.Wrap(err, "parsing channel")
	}

	creator := channel
	if creator, err = formatMessage(attrs.MustString("creator", &creator), m, r, eventData); err != nil {
		return false, errors.Wrap(err, "parsing creator")
	}

	canCreate, err := hasPerm(creator, twitch.ScopeClipsEdit)
	if err != nil {
		return false, errors.Wrap(err, "checking for required permission")
	}

	if !canCreate {
		return false, errors.Errorf("creator has not given %s permission", twitch.ScopeClipsEdit)
	}

	tc, err := tcGetter(creator)
	if err != nil {
		return false, errors.Wrapf(err, "getting Twitch client for %q", creator)
	}

	clipInfo, err := tc.CreateClip(context.TODO(), channel, attrs.MustBool("add_delay", helpers.Ptr(false)))
	if err != nil {
		return false, errors.Wrap(err, "creating clip")
	}

	eventData.Set("create_clip_slug", clipInfo.ID)
	eventData.Set("create_clip_edit_url", clipInfo.EditURL)
	return false, nil
}

func (actor) IsAsync() bool { return false }

func (actor) Name() string { return actorName }

func (actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "channel", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "creator", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "add_delay", Type: fieldcollection.SchemaFieldTypeBool}),
		fieldcollection.MustHaveNoUnknowFields,
		helpers.SchemaValidateTemplateField(tplValidator, "channel", "creator"),
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	return nil
}

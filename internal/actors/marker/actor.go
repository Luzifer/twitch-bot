// Package marker contains an actor to create markers on the current
// running stream
package marker

import (
	"context"
	"fmt"
	"strings"

	"github.com/Luzifer/go_helpers/v2/fieldcollection"
	"github.com/Luzifer/twitch-bot/v3/internal/helpers"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
	"gopkg.in/irc.v4"
)

const actorName = "marker"

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
		Description: "Creates a marker on the currently running stream of the given channel. The marker will be created on behalf of the channel owner and requires matching scope.",
		Name:        "Create Marker",
		Type:        actorName,
		Fields: []plugins.ActionDocumentationField{
			{
				Description:     "Channel to create the marker in, defaults to the channel of the event / message",
				Key:             "channel",
				Name:            "Channel",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Description:     "Description of the marker to create (up to 140 chars)",
				Key:             "description",
				Name:            "Description",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	return nil
}

type actor struct{}

func (actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	channel := plugins.DeriveChannel(m, eventData)
	if channel, err = formatMessage(attrs.MustString("channel", &channel), m, r, eventData); err != nil {
		return false, fmt.Errorf("parsing channel: %w", err)
	}

	var description string
	if description, err = formatMessage(attrs.MustString("description", &description), m, r, eventData); err != nil {
		return false, fmt.Errorf("parsing description: %w", err)
	}

	channel = strings.TrimLeft(channel, "#")

	canCreate, err := hasPerm(channel, twitch.ScopeChannelManageBroadcast)
	if err != nil {
		return false, fmt.Errorf("checking for required permission: %w", err)
	}

	if !canCreate {
		return false, fmt.Errorf("creator has not given %s permission", twitch.ScopeChannelManageBroadcast)
	}

	tc, err := tcGetter(channel)
	if err != nil {
		return false, fmt.Errorf("getting Twitch client for %q: %w", channel, err)
	}

	if err = tc.CreateStreamMarker(context.TODO(), description); err != nil {
		return false, fmt.Errorf("creating marker: %w", err)
	}

	return false, nil
}

func (actor) IsAsync() bool { return false }

func (actor) Name() string { return actorName }

func (actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "channel", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "description", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.MustHaveNoUnknowFields,
		helpers.SchemaValidateTemplateField(tplValidator, "channel", "description"),
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	return nil
}

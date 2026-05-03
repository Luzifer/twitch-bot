// Package modchannel contains an actor to modify title / category of
// a channel
package modchannel

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

const actorName = "modchannel"

type actor struct{}

var (
	formatMessage plugins.MsgFormatter
	tcGetter      func(string) (*twitch.Client, error)
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	formatMessage = args.FormatMessage
	tcGetter = args.GetTwitchClientForChannel

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: "Update stream information",
		Name:        "Modify Stream",
		Type:        "modchannel",

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Channel to update",
				Key:             "channel",
				Name:            "Channel",
				Optional:        false,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "",
				Description:     "Category / Game to set (use `@1234` format to pass an explicit ID)",
				Key:             "game",
				Name:            "Game",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "",
				Description:     "Stream title to set",
				Key:             "title",
				Name:            "Title",
				Optional:        true,
				SupportTemplate: true,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
		},
	})

	return nil
}

func (actor) Execute(_ *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	var (
		game  = attrs.MustString("game", helpers.Ptr(""))
		title = attrs.MustString("title", helpers.Ptr(""))
	)

	if game == "" && title == "" {
		return false, nil
	}

	var updGame, updTitle *string

	channel, err := formatMessage(attrs.MustString("channel", nil), m, r, eventData)
	if err != nil {
		return false, fmt.Errorf("parsing channel: %w", err)
	}

	if game != "" {
		parsedGame, err := formatMessage(game, m, r, eventData)
		if err != nil {
			return false, fmt.Errorf("parsing game: %w", err)
		}

		updGame = &parsedGame
	}

	if title != "" {
		parsedTitle, err := formatMessage(title, m, r, eventData)
		if err != nil {
			return false, fmt.Errorf("parsing title: %w", err)
		}

		updTitle = &parsedTitle
	}

	twitchClient, err := tcGetter(strings.TrimLeft(channel, "#"))
	if err != nil {
		return false, fmt.Errorf("getting Twitch client: %w", err)
	}

	if err = twitchClient.ModifyChannelInformation(
		context.Background(),
		strings.TrimLeft(channel, "#"),
		updGame,
		updTitle,
	); err != nil {
		return false, fmt.Errorf("updating channel info: %w", err)
	}

	return false, nil
}

func (actor) IsAsync() bool { return false }
func (actor) Name() string  { return actorName }

func (actor) Validate(tplValidator plugins.TemplateValidatorFunc, attrs *fieldcollection.FieldCollection) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.MustHaveField(fieldcollection.SchemaField{Name: "channel", NonEmpty: true, Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "game", Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.CanHaveField(fieldcollection.SchemaField{Name: "title", Type: fieldcollection.SchemaFieldTypeString}),
		fieldcollection.MustHaveNoUnknowFields,
		helpers.SchemaValidateTemplateField(tplValidator, "channel", "game", "title"),
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	return nil
}

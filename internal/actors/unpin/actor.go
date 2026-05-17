// Package unpin contains an actor to unpin any currently pinned message
package unpin

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "unpin"

// Actor implements the actor interface
type Actor struct{}

var (
	hasPerm  plugins.ChannelPermissionCheckFunc
	tcGetter func(string) (*twitch.Client, error)
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	hasPerm = args.HasPermissionForChannel
	tcGetter = args.GetTwitchClientForChannel

	args.RegisterActor(actorName, func() plugins.Actor { return &Actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: `Unpins any currently pinned message in the channel`,
		Name:        "Unpin Message",
		Type:        actorName,
	})

	return nil
}

// Execute implements the actor interface
func (Actor) Execute(
	_ *irc.Client,
	m *irc.Message,
	_ *plugins.Rule,
	eventData *fieldcollection.FieldCollection,
	_ *fieldcollection.FieldCollection,
) (preventCooldown bool, err error) {
	channel := strings.TrimLeft(plugins.DeriveChannel(m, eventData), "#")

	canModerate, err := hasPerm(channel, twitch.ScopeModeratorManageChatMessages)
	if err != nil {
		return false, fmt.Errorf("checking for required permission: %w", err)
	}

	if !canModerate {
		return false, fmt.Errorf("creator has not given %s permission", twitch.ScopeModeratorManageChatMessages)
	}

	twitchClient, err := tcGetter(channel)
	if err != nil {
		return false, fmt.Errorf("getting Twitch client: %w", err)
	}

	pin, err := twitchClient.GetPinnedChatMessage(context.TODO(), channel)
	if err != nil {
		if errors.Is(err, twitch.ErrNoPinnedChatMessage) {
			// Nothing to unpin
			return false, nil
		}

		return false, fmt.Errorf("getting pinned message: %w", err)
	}

	if err = twitchClient.UnpinChatMessage(context.TODO(), channel, pin.MessageID); err != nil {
		return false, fmt.Errorf("unpinning message: %w", err)
	}

	return false, nil
}

// IsAsync implements the actor interface
func (Actor) IsAsync() bool { return false }

// Name implements the actor interface
func (Actor) Name() string { return actorName }

// Validate implements the actor interface
func (Actor) Validate(
	_ plugins.TemplateValidatorFunc,
	attrs *fieldcollection.FieldCollection,
) (err error) {
	if err = attrs.ValidateSchema(
		fieldcollection.MustHaveNoUnknowFields,
	); err != nil {
		return fmt.Errorf("validating attributes: %w", err)
	}

	return nil
}

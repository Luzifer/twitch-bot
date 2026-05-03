// Package clipdetector contains an actor to detect clip links in a
// message and populate a template variable
package clipdetector

import (
	"context"
	"fmt"
	"regexp"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/internal/actors/linkdetector"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "clipdetector"

// Actor implements the actor interface
type Actor struct{}

var (
	botTwitchClient func() *twitch.Client
	clipIDScanner   = regexp.MustCompile(`(?:clips\.twitch\.tv|www\.twitch\.tv/[^/]*/clip)/([A-Za-z0-9_-]+)`)
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	botTwitchClient = args.GetTwitchClient

	args.RegisterActor(actorName, func() plugins.Actor { return &Actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: `Scans for clip-links in the message and adds the "clips" field to the event data`,
		Name:        "Scan for Clips",
		Type:        actorName,
	})

	return nil
}

// Execute implements the actor interface
func (Actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData *fieldcollection.FieldCollection, attrs *fieldcollection.FieldCollection) (preventCooldown bool, err error) {
	if eventData.HasAll("clips") {
		// We already detected clips, lets not do it again
		return false, nil
	}

	// In case the link detector did not run before, lets run it now
	if preventCooldown, err = (linkdetector.Actor{}).Execute(c, m, r, eventData, attrs); err != nil {
		return preventCooldown, fmt.Errorf("detecting links: %w", err)
	}

	links, err := eventData.StringSlice("links")
	if err != nil {
		return false, fmt.Errorf("getting links data: %w", err)
	}

	var clips []twitch.ClipInfo
	for _, link := range links {
		clipIDMatch := clipIDScanner.FindStringSubmatch(link)
		if clipIDMatch == nil {
			continue
		}

		clipInfo, err := botTwitchClient().GetClipByID(context.Background(), clipIDMatch[1])
		if err != nil {
			return false, fmt.Errorf("getting clip info: %w", err)
		}

		clips = append(clips, clipInfo)
	}

	eventData.Set("clips", clips)
	return false, nil
}

// IsAsync implements the actor interface
func (Actor) IsAsync() bool { return false }

// Name implements the actor interface
func (Actor) Name() string { return actorName }

// Validate implements the actor interface
func (Actor) Validate(plugins.TemplateValidatorFunc, *fieldcollection.FieldCollection) error {
	return nil
}

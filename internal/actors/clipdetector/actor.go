package clipdetector

import (
	"context"
	"regexp"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v3/internal/actors/linkdetector"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "clipdetector"

var (
	botTwitchClient *twitch.Client
	clipIDScanner   = regexp.MustCompile(`(?:clips\.twitch\.tv|www\.twitch\.tv/[^/]*/clip)/([A-Za-z0-9_-]+)`)
)

func Register(args plugins.RegistrationArguments) error {
	botTwitchClient = args.GetTwitchClient()

	args.RegisterActor(actorName, func() plugins.Actor { return &Actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: `Scans for clip-links in the message and adds the "clips" field to the event data`,
		Name:        "Scan for Clips",
		Type:        actorName,
	})

	return nil
}

type Actor struct{}

func (Actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	if eventData.HasAll("clips") {
		// We already detected clips, lets not do it again
		return false, nil
	}

	// In case the link detector did not run before, lets run it now
	if preventCooldown, err = (linkdetector.Actor{}).Execute(c, m, r, eventData, attrs); err != nil {
		return preventCooldown, errors.Wrap(err, "detecting links")
	}

	links, err := eventData.StringSlice("links")
	if err != nil {
		return false, errors.Wrap(err, "getting links data")
	}

	var clips []twitch.ClipInfo
	for _, link := range links {
		clipIDMatch := clipIDScanner.FindStringSubmatch(link)
		if clipIDMatch == nil {
			continue
		}

		clipInfo, err := botTwitchClient.GetClipByID(context.Background(), clipIDMatch[1])
		if err != nil {
			return false, errors.Wrap(err, "getting clip info")
		}

		clips = append(clips, clipInfo)
	}

	eventData.Set("clips", clips)
	return false, nil
}

func (Actor) IsAsync() bool { return false }

func (Actor) Name() string { return actorName }

func (Actor) Validate(plugins.TemplateValidatorFunc, *plugins.FieldCollection) error { return nil }

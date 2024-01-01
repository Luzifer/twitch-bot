// Package linkprotect contains an actor to prevent chatters from
// posting certain links
package linkprotect

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/internal/actors/clipdetector"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const actorName = "linkprotect"

var (
	botTwitchClient *twitch.Client
	clipLink        = regexp.MustCompile(`.*(?:clips\.twitch\.tv|www\.twitch\.tv/[^/]*/clip)/.*`)
	ptrBoolFalse    = func(v bool) *bool { return &v }(false)
	ptrStringEmpty  = func(v string) *string { return &v }("")
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	botTwitchClient = args.GetTwitchClient()

	args.RegisterActor(actorName, func() plugins.Actor { return &actor{} })

	args.RegisterActorDocumentation(plugins.ActionDocumentation{
		Description: `Uses link- and clip-scanner to detect links / clips and applies link protection as defined`,
		Name:        "Enforce Link-Protection",
		Type:        actorName,

		Fields: []plugins.ActionDocumentationField{
			{
				Default:         "",
				Description:     "Allowed links (if any is specified all non matching links will cause enforcement action, link must contain any of these strings)",
				Key:             "allowed_links",
				Name:            "Allowed Links",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeStringSlice,
			},
			{
				Default:         "",
				Description:     "Disallowed links (if any is specified all non matching links will not cause enforcement action, link must contain any of these strings)",
				Key:             "disallowed_links",
				Name:            "Disallowed Links",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeStringSlice,
			},
			{
				Default:         "",
				Description:     "Allowed clip channels (if any is specified clips of all other channels will cause enforcement action, clip-links will be ignored in link-protection when this is used)",
				Key:             "allowed_clip_channels",
				Name:            "Allowed Clip Channels",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeStringSlice,
			},
			{
				Default:         "",
				Description:     "Disallowed clip channels (if any is specified clips of all other channels will not cause enforcement action, clip-links will be ignored in link-protection when this is used)",
				Key:             "disallowed_clip_channels",
				Name:            "Disallowed Clip Channels",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeStringSlice,
			},
			{
				Default:         "",
				Description:     "Enforcement action to take when disallowed link / clip is detected (ban, delete, duration-value i.e. 1m)",
				Key:             "action",
				Name:            "Action",
				Optional:        false,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "",
				Description:     "Reason why the enforcement action was taken",
				Key:             "reason",
				Name:            "Reason",
				Optional:        false,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeString,
			},
			{
				Default:         "false",
				Description:     "Stop rule execution when action is applied (i.e. not to post a message after a ban for spam links)",
				Key:             "stop_on_action",
				Name:            "Stop on Action",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeBool,
			},
			{
				Default:         "false",
				Description:     "Stop rule execution when no action is applied (i.e. not to post a message when no enforcement action is taken)",
				Key:             "stop_on_no_action",
				Name:            "Stop on no Action",
				Optional:        true,
				SupportTemplate: false,
				Type:            plugins.ActionDocumentationFieldTypeBool,
			},
		},
	})

	return nil
}

type (
	actor struct{}

	verdict uint
)

const (
	verdictAllFine verdict = iota
	verdictMisbehave
)

//nolint:gocyclo // Minimum over the limit, makes no sense to split
func (a actor) Execute(c *irc.Client, m *irc.Message, r *plugins.Rule, eventData *plugins.FieldCollection, attrs *plugins.FieldCollection) (preventCooldown bool, err error) {
	// In case the clip detector did not run before, lets run it now
	if preventCooldown, err = (clipdetector.Actor{}).Execute(c, m, r, eventData, attrs); err != nil {
		return preventCooldown, errors.Wrap(err, "detecting links / clips")
	}

	links, err := eventData.StringSlice("links")
	if err != nil {
		return preventCooldown, errors.Wrap(err, "getting links from event")
	}

	if len(links) == 0 {
		// If there are no links there is nothing to protect and there
		// are also no clips as they are parsed from the links
		if attrs.MustBool("stop_on_no_action", ptrBoolFalse) {
			return false, plugins.ErrStopRuleExecution
		}
		return false, nil
	}

	clipsInterface, err := eventData.Any("clips")
	if err != nil {
		return preventCooldown, errors.Wrap(err, "getting clips from event")
	}
	clips, ok := clipsInterface.([]twitch.ClipInfo)
	if !ok {
		return preventCooldown, errors.New("invalid data-type in clips")
	}

	if a.check(links, clips, attrs) == verdictAllFine {
		if attrs.MustBool("stop_on_no_action", ptrBoolFalse) {
			return false, plugins.ErrStopRuleExecution
		}
		return false, nil
	}

	// That message misbehaved so we need to punish them
	switch lt := attrs.MustString("action", ptrStringEmpty); lt {
	case "ban":
		if err = botTwitchClient.BanUser(
			context.Background(),
			plugins.DeriveChannel(m, eventData),
			strings.TrimLeft(plugins.DeriveUser(m, eventData), "@"),
			0,
			attrs.MustString("reason", ptrStringEmpty),
		); err != nil {
			return false, errors.Wrap(err, "executing user ban")
		}

	case "delete":
		msgID, ok := m.Tags["id"]
		if !ok || msgID == "" {
			return false, errors.New("found no mesage id")
		}

		if err = botTwitchClient.DeleteMessage(
			context.Background(),
			plugins.DeriveChannel(m, eventData),
			msgID,
		); err != nil {
			return false, errors.Wrap(err, "deleting message")
		}

	default:
		to, err := time.ParseDuration(lt)
		if err != nil {
			return false, errors.Wrap(err, "parsing punishment level")
		}

		if err = botTwitchClient.BanUser(
			context.Background(),
			plugins.DeriveChannel(m, eventData),
			strings.TrimLeft(plugins.DeriveUser(m, eventData), "@"),
			to,
			attrs.MustString("reason", ptrStringEmpty),
		); err != nil {
			return false, errors.Wrap(err, "executing user ban")
		}
	}

	if attrs.MustBool("stop_on_action", ptrBoolFalse) {
		return false, plugins.ErrStopRuleExecution
	}

	return false, nil
}

func (actor) IsAsync() bool { return false }

func (actor) Name() string { return actorName }

func (actor) Validate(_ plugins.TemplateValidatorFunc, attrs *plugins.FieldCollection) error {
	if v, err := attrs.String("action"); err != nil || v == "" {
		return errors.New("action must be non-empty string")
	}

	if v, err := attrs.String("reason"); err != nil || v == "" {
		return errors.New("reason must be non-empty string")
	}

	if len(attrs.MustStringSlice("allowed_links"))+
		len(attrs.MustStringSlice("disallowed_links"))+
		len(attrs.MustStringSlice("allowed_clip_channels"))+
		len(attrs.MustStringSlice("disallowed_clip_channels")) == 0 {
		return errors.New("no conditions are provided")
	}

	return nil
}

func (a actor) check(links []string, clips []twitch.ClipInfo, attrs *plugins.FieldCollection) (v verdict) {
	hasClipDefinition := len(attrs.MustStringSlice("allowed_clip_channels"))+len(attrs.MustStringSlice("disallowed_clip_channels")) > 0

	if v = a.checkLinkDenied(attrs.MustStringSlice("disallowed_links"), links, hasClipDefinition); v == verdictMisbehave {
		return verdictMisbehave
	}

	if v = a.checkAllLinksAllowed(attrs.MustStringSlice("allowed_links"), links, hasClipDefinition); v == verdictMisbehave {
		return verdictMisbehave
	}

	if v = a.checkClipChannelDenied(attrs.MustStringSlice("disallowed_clip_channels"), clips); v == verdictMisbehave {
		return verdictMisbehave
	}

	if v = a.checkAllClipChannelsAllowed(attrs.MustStringSlice("allowed_clip_channels"), clips); v == verdictMisbehave {
		return verdictMisbehave
	}

	return verdictAllFine
}

func (actor) checkAllClipChannelsAllowed(allowList []string, clips []twitch.ClipInfo) verdict {
	if len(allowList) == 0 {
		// We're not explicitly allowing clip-channels, this method is a no-op
		return verdictAllFine
	}

	allAllowed := true
	for _, clip := range clips {
		clipAllowed := false
		for _, allowed := range allowList {
			if strings.EqualFold(clip.BroadcasterName, allowed) {
				clipAllowed = true
			}
		}

		allAllowed = allAllowed && clipAllowed
	}

	if allAllowed {
		// All clips are fine
		return verdictAllFine
	}

	// Some clips are not fine
	return verdictMisbehave
}

func (actor) checkClipChannelDenied(denyList []string, clips []twitch.ClipInfo) verdict {
	for _, clip := range clips {
		for _, denied := range denyList {
			if strings.EqualFold(clip.BroadcasterName, denied) {
				return verdictMisbehave
			}
		}
	}

	return verdictAllFine
}

//revive:disable-next-line:flag-parameter
func (actor) checkAllLinksAllowed(allowList, links []string, autoAllowClipLinks bool) verdict {
	if len(allowList) == 0 {
		// We're not explicitly allowing links, this method is a no-op
		return verdictAllFine
	}

	allAllowed := true
	for _, link := range links {
		if autoAllowClipLinks && clipLink.MatchString(link) {
			// The default is "true", so we don't change that in this case
			// as the expression would be `allowList && true` which is BS
			continue
		}

		var linkAllowed bool
		for _, allowed := range allowList {
			linkAllowed = linkAllowed || strings.Contains(strings.ToLower(link), strings.ToLower(allowed))
		}

		allAllowed = allAllowed && linkAllowed
	}

	if allAllowed {
		// All links are fine
		return verdictAllFine
	}

	// Some links are not fine
	return verdictMisbehave
}

//revive:disable-next-line:flag-parameter
func (actor) checkLinkDenied(denyList, links []string, ignoreClipLinks bool) verdict {
	for _, link := range links {
		if ignoreClipLinks && clipLink.MatchString(link) {
			// We have special directives for clips so we ignore clip-links
			continue
		}

		for _, denied := range denyList {
			if strings.Contains(strings.ToLower(link), strings.ToLower(denied)) {
				// Well, that link is definitely not allowed
				return verdictMisbehave
			}
		}
	}

	return verdictAllFine
}

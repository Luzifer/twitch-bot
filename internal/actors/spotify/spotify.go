// Package spotify contains an actor to query the current playing
// track for a channel with authorized spotify account
package spotify

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/Luzifer/twitch-bot/v3/pkg/database"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

const (
	actorName = "spotify"
)

var (
	baseURL         *url.URL
	db              database.Connector
	getModuleConfig plugins.ModuleConfigGetterFunc
)

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) (err error) {
	if baseURL, err = url.Parse(args.GetBaseURL()); err != nil {
		return fmt.Errorf("parsing base-url: %w", err)
	}

	db = args.GetDatabaseConnector()
	getModuleConfig = args.GetModuleConfigForChannel

	args.RegisterTemplateFunction(
		"spotifyCurrentPlaying",
		plugins.GenericTemplateFunctionGetter(getCurrentTrackForChannel),
		plugins.TemplateFuncDocumentation{
			Name:        "spotifyCurrentPlaying",
			Description: "Retrieves the current playing track for the given channel",
			Syntax:      "spotifyCurrentPlaying <channel>",
			Example: &plugins.TemplateFuncDocumentationExample{
				MatchMessage:   "^!spotify",
				MessageContent: "!spotify",
				Template:       "{{ spotifyCurrentPlaying .channel }}",
				FakedOutput:    "Beast in Black - Die By The Blade",
			},
		})

	if err = args.RegisterAPIRoute(plugins.HTTPRouteRegistrationArgs{
		Description:       "Starts authorization of an Spotify Account for a {channel}",
		HandlerFunc:       handleStartAuth,
		Method:            http.MethodGet,
		Module:            actorName,
		Name:              "Authorize Spotify Account",
		Path:              "/{channel}",
		RequiresWriteAuth: false,
		ResponseType:      plugins.HTTPRouteResponseTypeTextPlain,
		RouteParams: []plugins.HTTPRouteParamDocumentation{
			{
				Description: "Channel to authorize the Account for",
				Name:        "channel",
			},
		},
	}); err != nil {
		return fmt.Errorf("registering API route: %w", err)
	}

	return nil
}

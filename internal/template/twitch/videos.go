package twitch

import (
	"context"
	"fmt"
	"strings"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

func init() {
	regFn = append(
		regFn,
		tplTwitchCurrentVOD,
	)
}

func tplTwitchCurrentVOD(args plugins.RegistrationArguments) {
	args.RegisterTemplateFunction("currentVOD", plugins.GenericTemplateFunctionGetter(func(username string) (string, error) {
		si, err := args.GetTwitchClient().GetCurrentStreamInfo(context.Background(), strings.TrimLeft(username, "#"))
		if err != nil {
			return "", fmt.Errorf("getting stream info: %w", err)
		}

		vids, err := args.GetTwitchClient().GetVideos(context.TODO(), twitch.GetVideoOpts{
			UserID: si.UserID,
		})
		if err != nil {
			return "", fmt.Errorf("getting videos: %w", err)
		}

		for _, v := range vids {
			if v.StreamID == nil || *v.StreamID != si.ID {
				continue
			}

			return v.URL, nil
		}

		return "", fmt.Errorf("no matching VOD found")
	}), plugins.TemplateFuncDocumentation{
		Description: "Returns the VOD of the currently running stream in the given channel (causes an error if no current stream / VOD is found)",
		Syntax:      "currentVOD <username>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ currentVOD .channel }}`,
			FakedOutput: "https://www.twitch.tv/videos/123456789",
		},
	})
}

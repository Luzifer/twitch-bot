package twitch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Luzifer/twitch-bot/v3/plugins"
	"github.com/pkg/errors"
)

func init() {
	regFn = append(
		regFn,
		tplTwitchRecentGame,
		tplTwitchRecentTitle,
		tplTwitchStreamIsLive,
		tplTwitchStreamUptime,
	)
}

func tplTwitchRecentGame(args plugins.RegistrationArguments) {
	args.RegisterTemplateFunction("recentGame", plugins.GenericTemplateFunctionGetter(func(username string, v ...string) (string, error) {
		game, _, err := args.GetTwitchClient().GetRecentStreamInfo(context.Background(), strings.TrimLeft(username, "#"))
		if len(v) > 0 && (err != nil || game == "") {
			return v[0], nil //nolint:nilerr // This is a default fallback
		}

		return game, errors.Wrap(err, "getting stream info")
	}), plugins.TemplateFuncDocumentation{
		Description: "Returns the last played game name of the specified user (see shoutout example) or the `fallback` if the game could not be fetched. If no fallback was supplied the message will fail and not be sent.",
		Syntax:      "recentGame <username> [fallback]",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ recentGame "luziferus" "none" }} - {{ recentGame "thisuserdoesnotexist123" "none" }}`,
			FakedOutput: "Metro Exodus - none",
		},
	})
}

func tplTwitchRecentTitle(args plugins.RegistrationArguments) {
	args.RegisterTemplateFunction("recentTitle", plugins.GenericTemplateFunctionGetter(func(username string, v ...string) (string, error) {
		_, title, err := args.GetTwitchClient().GetRecentStreamInfo(context.Background(), strings.TrimLeft(username, "#"))
		if len(v) > 0 && (err != nil || title == "") {
			return v[0], nil //nolint:nilerr // This is a default fallback
		}

		return title, errors.Wrap(err, "getting stream info")
	}), plugins.TemplateFuncDocumentation{
		Description: "Returns the last stream title of the specified user or the `fallback` if the title could not be fetched. If no fallback was supplied the message will fail and not be sent.",
		Syntax:      "recentTitle <username> [fallback]",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ recentGame "luziferus" "none" }} - {{ recentGame "thisuserdoesnotexist123" "none" }}`,
			FakedOutput: "Die Oper haben wir überlebt, mal sehen was uns sonst noch alles töten möchte… - none",
		},
	})
}

func tplTwitchStreamIsLive(args plugins.RegistrationArguments) {
	args.RegisterTemplateFunction("streamIsLive", plugins.GenericTemplateFunctionGetter(func(username string) bool {
		_, err := args.GetTwitchClient().GetCurrentStreamInfo(context.Background(), strings.TrimLeft(username, "#"))
		return err == nil
	}), plugins.TemplateFuncDocumentation{
		Description: "Check whether a given channel is currently live",
		Syntax:      "streamIsLive <username>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ streamIsLive "luziferus" }}`,
			FakedOutput: "true",
		},
	})
}

func tplTwitchStreamUptime(args plugins.RegistrationArguments) {
	args.RegisterTemplateFunction("streamUptime", plugins.GenericTemplateFunctionGetter(func(username string) (time.Duration, error) {
		si, err := args.GetTwitchClient().GetCurrentStreamInfo(context.Background(), strings.TrimLeft(username, "#"))
		if err != nil {
			return 0, fmt.Errorf("getting stream info: %w", err)
		}
		return time.Since(si.StartedAt), nil
	}), plugins.TemplateFuncDocumentation{
		Description: "Returns the duration the stream is online (causes an error if no current stream is found)",
		Syntax:      "streamUptime <username>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ formatDuration (streamUptime "luziferus") "hours" "minutes" "" }}`,
			FakedOutput: "3 hours, 56 minutes",
		},
	})
}

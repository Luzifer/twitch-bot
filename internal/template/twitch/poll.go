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
		tplTwitchLastPoll,
	)
}

func tplTwitchLastPoll(args plugins.RegistrationArguments) {
	args.RegisterTemplateFunction("lastPoll", plugins.GenericTemplateFunctionGetter(func(username string) (*twitch.PollInfo, error) {
		hasPollAccess, err := args.HasAnyPermissionForChannel(username, twitch.ScopeChannelReadPolls, twitch.ScopeChannelManagePolls)
		if err != nil {
			return nil, fmt.Errorf("checking read-poll-permission: %w", err)
		}

		if !hasPollAccess {
			return nil, fmt.Errorf("not authorized to read polls for channel %s", username)
		}

		tc, err := args.GetTwitchClientForChannel(strings.TrimLeft(username, "#"))
		if err != nil {
			return nil, fmt.Errorf("getting twitch client for user: %w", err)
		}

		poll, err := tc.GetLatestPoll(context.Background(), strings.TrimLeft(username, "#"))
		if err != nil {
			return poll, fmt.Errorf("getting last poll: %w", err)
		}

		return poll, nil
	}), plugins.TemplateFuncDocumentation{
		Description: "Gets the last (currently running or archived) poll for the given channel (the channel must have given extended permission for poll access!)",
		Syntax:      "lastPoll <channel>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `Last Poll: {{ (lastPoll .channel).Title }}`,
			FakedOutput: "Last Poll: Und wie siehts im Template aus?",
		},
		Remarks: "See schema of returned object in [`pkg/twitch/polls.go#L13`](https://github.com/Luzifer/twitch-bot/blob/master/pkg/twitch/polls.go#L13)",
	})
}

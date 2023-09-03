package twitch

import (
	"context"
	"strings"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
	"github.com/pkg/errors"
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
			return nil, errors.Wrap(err, "checking read-poll-permission")
		}

		if !hasPollAccess {
			return nil, errors.Errorf("not authorized to read polls for channel %s", username)
		}

		tc, err := args.GetTwitchClientForChannel(strings.TrimLeft(username, "#"))
		if err != nil {
			return nil, errors.Wrap(err, "getting twitch client for user")
		}

		poll, err := tc.GetLatestPoll(context.Background(), strings.TrimLeft(username, "#"))
		return poll, errors.Wrap(err, "getting last poll")
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

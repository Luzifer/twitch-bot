package subscriber

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

var (
	permCheckFn plugins.ChannelPermissionCheckFunc
	tcGetter    func(string) (*twitch.Client, error)
)

func Register(args plugins.RegistrationArguments) error {
	permCheckFn = args.HasPermissionForChannel
	tcGetter = args.GetTwitchClientForChannel

	args.RegisterTemplateFunction("subCount", plugins.GenericTemplateFunctionGetter(subCount), plugins.TemplateFuncDocumentation{
		Description: "Returns the number of subscribers (accounts) currently subscribed to the given channel",
		Syntax:      "subCount <channel>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ subCount "luziferus" }}`,
			FakedOutput: "26",
		},
	})

	args.RegisterTemplateFunction("subPoints", plugins.GenericTemplateFunctionGetter(subPoints), plugins.TemplateFuncDocumentation{
		Description: "Returns the number of sub-points currently given through the T1 / T2 / T3 subscriptions to the given channel",
		Syntax:      "subPoints <channel>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ subPoints "luziferus" }}`,
			FakedOutput: "26",
		},
	})

	return nil
}

func getSubInfo(broadcasterName string) (subCount, subPoints int64, err error) {
	broadcasterName = strings.TrimLeft(broadcasterName, "#")

	ok, err := permCheckFn(broadcasterName, twitch.ScopeChannelReadSubscriptions)
	if err != nil {
		return 0, 0, errors.Wrap(err, "checking for channel permissions")
	}

	if !ok {
		return 0, 0, errors.Errorf("channel %q is missing permission %s", broadcasterName, twitch.ScopeChannelReadSubscriptions)
	}

	tc, err := tcGetter(broadcasterName)
	if err != nil {
		return 0, 0, errors.Wrap(err, "getting channel twitch-client")
	}

	sc, sp, err := tc.GetBroadcasterSubscriptionCount(context.Background(), broadcasterName)
	return sc, sp, errors.Wrap(err, "fetching sub info")
}

func subCount(broadcasterName string) (int64, error) {
	sc, _, err := getSubInfo(broadcasterName)
	return sc, err
}

func subPoints(broadcasterName string) (int64, error) {
	_, sp, err := getSubInfo(broadcasterName)
	return sp, err
}

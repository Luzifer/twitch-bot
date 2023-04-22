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

	args.RegisterTemplateFunction("subCount", plugins.GenericTemplateFunctionGetter(subCount))
	args.RegisterTemplateFunction("subPoints", plugins.GenericTemplateFunctionGetter(subPoints))
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

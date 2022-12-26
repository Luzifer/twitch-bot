package main

import (
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

func init() {
	tplFuncs.Register("displayName", plugins.GenericTemplateFunctionGetter(func(username string, v ...string) (string, error) {
		displayName, err := twitchClient.GetDisplayNameForUser(strings.TrimLeft(username, "#"))
		if len(v) > 0 && (err != nil || displayName == "") {
			return v[0], nil
		}

		return displayName, err
	}))

	tplFuncs.Register("followAge", plugins.GenericTemplateFunctionGetter(func(from, to string) (time.Duration, error) {
		since, err := twitchClient.GetFollowDate(from, to)
		return time.Since(since), errors.Wrap(err, "getting follow date")
	}))
	tplFuncs.Register("followDate", plugins.GenericTemplateFunctionGetter(func(from, to string) (time.Time, error) { return twitchClient.GetFollowDate(from, to) }))

	tplFuncs.Register("recentGame", plugins.GenericTemplateFunctionGetter(func(username string, v ...string) (string, error) {
		game, _, err := twitchClient.GetRecentStreamInfo(strings.TrimLeft(username, "#"))
		if len(v) > 0 && (err != nil || game == "") {
			return v[0], nil
		}

		return game, err
	}))

	tplFuncs.Register("streamUptime", plugins.GenericTemplateFunctionGetter(func(username string) (time.Duration, error) {
		si, err := twitchClient.GetCurrentStreamInfo(strings.TrimLeft(username, "#"))
		if err != nil {
			return 0, err
		}
		return time.Since(si.StartedAt), nil
	}))
}

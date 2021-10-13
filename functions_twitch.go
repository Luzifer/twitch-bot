package main

import (
	"strings"
	"time"

	"github.com/Luzifer/twitch-bot/plugins"
)

func init() {
	tplFuncs.Register("displayName", plugins.GenericTemplateFunctionGetter(func(username string, v ...string) (string, error) {
		displayName, err := twitchClient.GetDisplayNameForUser(strings.TrimLeft(username, "#"))
		if len(v) > 0 && (err != nil || displayName == "") {
			return v[0], nil
		}

		return displayName, err
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

package main

import (
	"strings"

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

	tplFuncs.Register("recentGame", plugins.GenericTemplateFunctionGetter(func(username string, v ...string) (string, error) {
		game, _, err := twitchClient.GetRecentStreamInfo(strings.TrimLeft(username, "#"))
		if len(v) > 0 && (err != nil || game == "") {
			return v[0], nil
		}

		return game, err
	}))
}

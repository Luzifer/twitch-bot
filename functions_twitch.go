package main

import (
	"strings"
)

func init() {
	tplFuncs.Register("displayName", genericTemplateFunctionGetter(func(username string, v ...string) (string, error) {
		displayName, err := twitchClient.GetDisplayNameForUser(strings.TrimLeft(username, "#"))
		if len(v) > 0 && (err != nil || displayName == "") {
			return v[0], nil
		}

		return displayName, err
	}))

	tplFuncs.Register("recentGame", genericTemplateFunctionGetter(func(username string, v ...string) (string, error) {
		game, _, err := twitchClient.GetRecentStreamInfo(strings.TrimLeft(username, "#"))
		if len(v) > 0 && (err != nil || game == "") {
			return v[0], nil
		}

		return game, err
	}))
}

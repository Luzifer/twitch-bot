package main

import (
	"strings"
)

func init() {
	tplFuncs.Register("recentGame", genericTemplateFunctionGetter(func(username string, v ...string) (string, error) {
		game, _, err := twitch.GetRecentStreamInfo(strings.TrimLeft(username, "#"))
		if err != nil && len(v) > 0 {
			return v[0], nil
		}

		return game, err
	}))
}

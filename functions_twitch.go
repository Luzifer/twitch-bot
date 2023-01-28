package main

import (
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

func init() {
	tplFuncs.Register("displayName", plugins.GenericTemplateFunctionGetter(tplTwitchDisplayName))
	tplFuncs.Register("doesFollow", plugins.GenericTemplateFunctionGetter(tplTwitchDoesFollow))
	tplFuncs.Register("followAge", plugins.GenericTemplateFunctionGetter(tplTwitchFollowAge))
	tplFuncs.Register("followDate", plugins.GenericTemplateFunctionGetter(tplTwitchFollowDate))
	tplFuncs.Register("doesFollowLongerThan", plugins.GenericTemplateFunctionGetter(tplTwitchDoesFollowLongerThan))
	tplFuncs.Register("recentGame", plugins.GenericTemplateFunctionGetter(tplTwitchRecentGame))
	tplFuncs.Register("recentTitle", plugins.GenericTemplateFunctionGetter(tplTwitchRecentTitle))
	tplFuncs.Register("streamUptime", plugins.GenericTemplateFunctionGetter(tplTwitchStreamUptime))
}

func tplTwitchDisplayName(username string, v ...string) (string, error) {
	displayName, err := twitchClient.GetDisplayNameForUser(strings.TrimLeft(username, "#"))
	if len(v) > 0 && (err != nil || displayName == "") {
		return v[0], nil
	}

	return displayName, err
}

func tplTwitchDoesFollow(from, to string) (bool, error) {
	_, err := twitchClient.GetFollowDate(from, to)
	switch {
	case err == nil:
		return true, nil

	case errors.Is(err, twitch.ErrUserDoesNotFollow):
		return false, nil

	default:
		return false, errors.Wrap(err, "getting follow date")
	}
}

func tplTwitchFollowAge(from, to string) (time.Duration, error) {
	since, err := twitchClient.GetFollowDate(from, to)
	return time.Since(since), errors.Wrap(err, "getting follow date")
}

func tplTwitchFollowDate(from, to string) (time.Time, error) {
	return twitchClient.GetFollowDate(from, to)
}

func tplTwitchDoesFollowLongerThan(from, to string, t any) (bool, error) {
	var (
		age time.Duration
		err error
	)

	switch v := t.(type) {
	case int64:
		age = time.Duration(v) * time.Second

	case string:
		if age, err = time.ParseDuration(v); err != nil {
			return false, errors.Wrap(err, "parsing duration")
		}

	default:
		return false, errors.Errorf("unexpected input for duration %t", t)
	}

	fd, err := twitchClient.GetFollowDate(from, to)
	switch {
	case err == nil:
		return time.Since(fd) > age, nil

	case errors.Is(err, twitch.ErrUserDoesNotFollow):
		return false, nil

	default:
		return false, errors.Wrap(err, "getting follow date")
	}
}

func tplTwitchRecentGame(username string, v ...string) (string, error) {
	game, _, err := twitchClient.GetRecentStreamInfo(strings.TrimLeft(username, "#"))
	if len(v) > 0 && (err != nil || game == "") {
		return v[0], nil
	}

	return game, err
}

func tplTwitchRecentTitle(username string, v ...string) (string, error) {
	_, title, err := twitchClient.GetRecentStreamInfo(strings.TrimLeft(username, "#"))
	if len(v) > 0 && (err != nil || title == "") {
		return v[0], nil
	}

	return title, err
}

func tplTwitchStreamUptime(username string) (time.Duration, error) {
	si, err := twitchClient.GetCurrentStreamInfo(strings.TrimLeft(username, "#"))
	if err != nil {
		return 0, err
	}
	return time.Since(si.StartedAt), nil
}

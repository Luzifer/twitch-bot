package main

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v3/internal/service/access"
	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

//nolint:funlen
func init() {
	tplFuncs.Register("displayName", plugins.GenericTemplateFunctionGetter(tplTwitchDisplayName), plugins.TemplateFuncDocumentation{
		Description: "Returns the display name the specified user set for themselves",
		Syntax:      "displayName <username> [fallback]",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ displayName "luziferus" }} - {{ displayName "notexistinguser" "foobar" }}`,
			FakedOutput: "Luziferus - foobar",
		},
	})

	tplFuncs.Register("doesFollowLongerThan", plugins.GenericTemplateFunctionGetter(tplTwitchDoesFollowLongerThan), plugins.TemplateFuncDocumentation{
		Description: "Returns whether `from` follows `to` for more than `duration`",
		Syntax:      "doesFollowLongerThan <from> <to> <duration>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ doesFollowLongerThan "tezrian" "luziferus" "168h" }}`,
			FakedOutput: "true",
		},
	})

	tplFuncs.Register("doesFollow", plugins.GenericTemplateFunctionGetter(tplTwitchDoesFollow), plugins.TemplateFuncDocumentation{
		Description: "Returns whether `from` follows `to`",
		Syntax:      "doesFollow <from> <to>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ doesFollow "tezrian" "luziferus" }}`,
			FakedOutput: "true",
		},
	})

	tplFuncs.Register("followAge", plugins.GenericTemplateFunctionGetter(tplTwitchFollowAge), plugins.TemplateFuncDocumentation{
		Description: "Looks up when `from` followed `to` and returns the duration between then and now",
		Syntax:      "followAge <from> <to>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ followAge "tezrian" "luziferus" }}`,
			FakedOutput: "15004h14m59.116620989s",
		},
	})

	tplFuncs.Register("followDate", plugins.GenericTemplateFunctionGetter(tplTwitchFollowDate), plugins.TemplateFuncDocumentation{
		Description: "Looks up when `from` followed `to`",
		Syntax:      "followDate <from> <to>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ followDate "tezrian" "luziferus" }}`,
			FakedOutput: "2021-04-10 16:07:07 +0000 UTC",
		},
	})

	tplFuncs.Register("idForUsername", plugins.GenericTemplateFunctionGetter(tplTwitchIDForUsername), plugins.TemplateFuncDocumentation{
		Description: "Returns the user-id for the given username",
		Syntax:      "idForUsername <username>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ idForUsername "twitch" }}`,
			FakedOutput: "12826",
		},
	})

	tplFuncs.Register("lastPoll", plugins.GenericTemplateFunctionGetter(tplTwitchLastPoll), plugins.TemplateFuncDocumentation{
		Description: "Gets the last (currently running or archived) poll for the given channel (the channel must have given extended permission for poll access!)",
		Syntax:      "lastPoll <channel>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `Last Poll: {{ (lastPoll .channel).Title }}`,
			FakedOutput: "Last Poll: Und wie siehts im Template aus?",
		},
		Remarks: "See schema of returned object in [`pkg/twitch/polls.go#L13`](https://github.com/Luzifer/twitch-bot/blob/master/pkg/twitch/polls.go#L13)",
	})

	tplFuncs.Register("profileImage", plugins.GenericTemplateFunctionGetter(tplTwitchProfileImage), plugins.TemplateFuncDocumentation{
		Description: "Gets the URL of the given users profile image",
		Syntax:      "profileImage <username>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ profileImage .username }}`,
			FakedOutput: "https://static-cdn.jtvnw.net/jtv_user_pictures/[...].png",
		},
	})

	tplFuncs.Register("recentGame", plugins.GenericTemplateFunctionGetter(tplTwitchRecentGame), plugins.TemplateFuncDocumentation{
		Description: "Returns the last played game name of the specified user (see shoutout example) or the `fallback` if the game could not be fetched. If no fallback was supplied the message will fail and not be sent.",
		Syntax:      "recentGame <username> [fallback]",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ recentGame "luziferus" "none" }} - {{ recentGame "thisuserdoesnotexist123" "none" }}`,
			FakedOutput: "Metro Exodus - none",
		},
	})

	tplFuncs.Register("recentTitle", plugins.GenericTemplateFunctionGetter(tplTwitchRecentTitle), plugins.TemplateFuncDocumentation{
		Description: "Returns the last stream title of the specified user or the `fallback` if the title could not be fetched. If no fallback was supplied the message will fail and not be sent.",
		Syntax:      "recentTitle <username> [fallback]",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ recentGame "luziferus" "none" }} - {{ recentGame "thisuserdoesnotexist123" "none" }}`,
			FakedOutput: "Die Oper haben wir überlebt, mal sehen was uns sonst noch alles töten möchte… - none",
		},
	})

	tplFuncs.Register("streamUptime", plugins.GenericTemplateFunctionGetter(tplTwitchStreamUptime), plugins.TemplateFuncDocumentation{
		Description: "Returns the duration the stream is online (causes an error if no current stream is found)",
		Syntax:      "streamUptime <username>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ formatDuration (streamUptime "luziferus") "hours" "minutes" "" }}`,
			FakedOutput: "3 hours, 56 minutes",
		},
	})

	tplFuncs.Register("usernameForID", plugins.GenericTemplateFunctionGetter(tplTwitchUsernameForID), plugins.TemplateFuncDocumentation{
		Description: "Returns the current login name of an user-id",
		Syntax:      "usernameForID <user-id>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:    `{{ usernameForID "12826" }}`,
			FakedOutput: "twitch",
		},
	})
}

func tplTwitchDisplayName(username string, v ...string) (string, error) {
	displayName, err := twitchClient.GetDisplayNameForUser(strings.TrimLeft(username, "#"))
	if len(v) > 0 && (err != nil || displayName == "") {
		return v[0], nil //nolint:nilerr // Default value, no need to return error
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

func tplTwitchIDForUsername(username string) (string, error) {
	return twitchClient.GetIDForUsername(username)
}

func tplTwitchLastPoll(username string) (*twitch.PollInfo, error) {
	hasPollAccess, err := accessService.HasAnyPermissionForChannel(username, twitch.ScopeChannelReadPolls, twitch.ScopeChannelManagePolls)
	if err != nil {
		return nil, errors.Wrap(err, "checking read-poll-permission")
	}

	if !hasPollAccess {
		return nil, errors.Errorf("not authorized to read polls for channel %s", username)
	}

	tc, err := accessService.GetTwitchClientForChannel(strings.TrimLeft(username, "#"), access.ClientConfig{
		TwitchClient:       cfg.TwitchClient,
		TwitchClientSecret: cfg.TwitchClientSecret,
	})
	if err != nil {
		return nil, errors.Wrap(err, "getting twitch client for user")
	}

	poll, err := tc.GetLatestPoll(context.Background(), strings.TrimLeft(username, "#"))
	return poll, errors.Wrap(err, "getting last poll")
}

func tplTwitchProfileImage(username string) (string, error) {
	user, err := twitchClient.GetUserInformation(strings.TrimLeft(username, "#@"))
	if err != nil {
		return "", errors.Wrap(err, "getting user info")
	}

	return user.ProfileImageURL, nil
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

func tplTwitchUsernameForID(id string) (string, error) {
	return twitchClient.GetUsernameForID(context.Background(), id)
}

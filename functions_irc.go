package main

import (
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/pkg/twitch"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

func init() {
	tplFuncs.Register("arg", func(m *irc.Message, _ *plugins.Rule, _ *plugins.FieldCollection) interface{} {
		return func(arg int) (string, error) {
			msgParts := strings.Split(m.Trailing(), " ")
			if len(msgParts) <= arg {
				return "", errors.New("argument not found")
			}

			return msgParts[arg], nil
		}
	}, plugins.TemplateFuncDocumentation{
		Description: "Takes the message sent to the channel, splits by space and returns the Nth element",
		Syntax:      "arg <index>",
		Example: &plugins.TemplateFuncDocumentationExample{
			MessageContent: "!bsg @tester",
			Template:       `{{ arg 1 }} please refrain from BSG`,
			ExpectedOutput: `@tester please refrain from BSG`,
		},
	})

	tplFuncs.Register("chatterHasBadge", func(m *irc.Message, _ *plugins.Rule, _ *plugins.FieldCollection) interface{} {
		return func(badge string) bool {
			badges := twitch.ParseBadgeLevels(m)
			return badges.Has(badge)
		}
	}, plugins.TemplateFuncDocumentation{
		Description: "Checks whether chatter writing the current line has the given badge in the current channel",
		Syntax:      "chatterHasBadge <badge>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:       `{{ chatterHasBadge "moderator" }}`,
			ExpectedOutput: "true",
		},
	})

	tplFuncs.Register(
		"fixUsername",
		plugins.GenericTemplateFunctionGetter(func(username string) string { return strings.TrimLeft(username, "@#") }),
		plugins.TemplateFuncDocumentation{
			Description: "Ensures the username no longer contains the `@` or `#` prefix",
			Syntax:      "fixUsername <username>",
			Example: &plugins.TemplateFuncDocumentationExample{
				Template:       `{{ fixUsername .channel }} - {{ fixUsername "@luziferus" }}`,
				ExpectedOutput: "example - luziferus",
			},
		},
	)

	tplFuncs.Register("group", func(m *irc.Message, r *plugins.Rule, _ *plugins.FieldCollection) interface{} {
		return func(idx int, fallback ...string) (string, error) {
			fields := r.GetMatchMessage().FindStringSubmatch(m.Trailing())
			if len(fields) <= idx {
				return "", errors.New("group not found")
			}

			if fields[idx] == "" && len(fallback) > 0 {
				return fallback[0], nil
			}

			return fields[idx], nil
		}
	}, plugins.TemplateFuncDocumentation{
		Description: "Gets matching group specified by index from `match_message` regular expression, when `fallback` is defined, it is used when group has an empty match",
		Syntax:      "group <idx> [fallback]",
		Example: &plugins.TemplateFuncDocumentationExample{
			MatchMessage:   "!command ([0-9]+) ([a-z]+) ?([a-z]*)",
			MessageContent: "!command 12 test",
			Template:       `{{ group 2 "oops" }} - {{ group 3 "oops" }}`,
			ExpectedOutput: "test - oops",
		},
	})

	tplFuncs.Register(
		"mention",
		plugins.GenericTemplateFunctionGetter(func(username string) string { return "@" + strings.TrimLeft(username, "@#") }),
		plugins.TemplateFuncDocumentation{
			Description: "Strips username and converts into a mention",
			Syntax:      "mention <username>",
			Example: &plugins.TemplateFuncDocumentationExample{
				Template:       `{{ mention "@user" }} {{ mention "user" }} {{ mention "#user" }}`,
				ExpectedOutput: "@user @user @user",
			},
		},
	)

	tplFuncs.Register("tag", func(m *irc.Message, _ *plugins.Rule, _ *plugins.FieldCollection) interface{} {
		return func(tag string) string { return m.Tags[tag] }
	}, plugins.TemplateFuncDocumentation{
		Description: "Takes the message sent to the channel, returns the value of the tag specified",
		Syntax:      "tag <tagname>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:       `{{ tag "display-name" }}`,
			ExpectedOutput: "ExampleUser",
		},
	})
}

// Package userstate traces the bot state and provides template
// functions based on it
package userstate

import (
	"github.com/Luzifer/go_helpers/fieldcollection"
	"github.com/Luzifer/twitch-bot/v3/plugins"
	"github.com/pkg/errors"
	"gopkg.in/irc.v4"
)

var userState = newTwitchUserStateStore()

// Register provides the plugins.RegisterFunc
func Register(args plugins.RegistrationArguments) error {
	if err := args.RegisterRawMessageHandler(rawMessageHandler); err != nil {
		return errors.Wrap(err, "registering raw message handler")
	}

	args.RegisterTemplateFunction("botHasBadge", func(m *irc.Message, _ *plugins.Rule, fields *fieldcollection.FieldCollection) interface{} {
		return func(badge string) bool {
			state := userState.Get(plugins.DeriveChannel(m, fields))
			if state == nil {
				return false
			}
			return state.Badges.Has(badge)
		}
	}, plugins.TemplateFuncDocumentation{
		Description: "Checks whether bot has the given badge in the current channel",
		Syntax:      "botHasBadge <badge>",
		Example: &plugins.TemplateFuncDocumentationExample{
			Template:       `{{ botHasBadge "moderator" }}`,
			ExpectedOutput: "false",
		},
	})

	return nil
}

func rawMessageHandler(m *irc.Message) error {
	if m.Command != "USERSTATE" {
		return nil
	}

	state, err := parseTwitchUserState(m)
	if err != nil {
		return errors.Wrap(err, "parsing state")
	}

	userState.Set(plugins.DeriveChannel(m, nil), state)

	return nil
}

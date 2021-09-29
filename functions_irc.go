package main

import (
	"strings"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func init() {
	tplFuncs.Register("arg", func(m *irc.Message, r *plugins.Rule, fields map[string]interface{}) interface{} {
		return func(arg int) (string, error) {
			msgParts := strings.Split(m.Trailing(), " ")
			if len(msgParts) <= arg {
				return "", errors.New("argument not found")
			}

			return msgParts[arg], nil
		}
	})

	tplFuncs.Register("botHasBadge", func(m *irc.Message, r *plugins.Rule, fields map[string]interface{}) interface{} {
		return func(badge string) bool {
			state := botUserstate.Get(plugins.DeriveChannel(m, nil))
			if state == nil {
				return false
			}

			return state.Badges.Has(badge)
		}
	})

	tplFuncs.Register("fixUsername", plugins.GenericTemplateFunctionGetter(func(username string) string { return strings.TrimLeft(username, "@#") }))

	tplFuncs.Register("group", func(m *irc.Message, r *plugins.Rule, fields map[string]interface{}) interface{} {
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
	})

	tplFuncs.Register("tag", func(m *irc.Message, r *plugins.Rule, fields map[string]interface{}) interface{} {
		return func(tag string) string {
			s, _ := m.GetTag(tag)
			return s
		}
	})
}

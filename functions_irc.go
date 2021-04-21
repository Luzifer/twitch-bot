package main

import (
	"strings"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func init() {
	tplFuncs.Register("arg", func(m *irc.Message, r *rule, fields map[string]interface{}) interface{} {
		return func(arg int) (string, error) {
			msgParts := strings.Split(m.Trailing(), " ")
			if len(msgParts) <= arg {
				return "", errors.New("argument not found")
			}

			return msgParts[arg], nil
		}
	})

	tplFuncs.Register("fixUsername", genericTemplateFunctionGetter(func(username string) string { return strings.TrimLeft(username, "@#") }))

	tplFuncs.Register("group", func(m *irc.Message, r *rule, fields map[string]interface{}) interface{} {
		return func(idx int) (string, error) {
			fields := r.matchMessage.FindStringSubmatch(m.Trailing())
			if len(fields) <= idx {
				return "", errors.New("group not found")
			}

			return fields[idx], nil
		}
	})

	tplFuncs.Register("tag", func(m *irc.Message, r *rule, fields map[string]interface{}) interface{} {
		return func(tag string) string {
			s, _ := m.GetTag(tag)
			return s
		}
	})
}

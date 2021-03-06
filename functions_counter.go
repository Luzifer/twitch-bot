package main

import (
	"strings"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func init() {
	tplFuncs.Register("channelCounter", func(m *irc.Message, r *Rule, fields map[string]interface{}) interface{} {
		return func(name string) (string, error) {
			channel, ok := fields["channel"].(string)
			if !ok {
				return "", errors.New("channel not available")
			}

			return strings.Join([]string{channel, name}, ":"), nil
		}
	})

	tplFuncs.Register("counterValue", genericTemplateFunctionGetter(func(name string, _ ...string) int64 {
		return store.GetCounterValue(name)
	}))
}

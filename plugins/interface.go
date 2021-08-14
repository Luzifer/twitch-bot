package plugins

import "github.com/go-irc/irc"

type (
	MsgFormatter func(tplString string, m *irc.Message, r *Rule, fields map[string]interface{}) (string, error)
	RegisterFunc func() error
)

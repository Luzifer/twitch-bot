package main

import (
	"bytes"
	"strings"
	"text/template"
	"time"

	korvike "github.com/Luzifer/korvike/functions"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

var messageFunctions = korvike.GetFunctionMap()

func formatMessage(tplString string, m *irc.Message, fields map[string]interface{}) (string, error) {
	tpl, err := template.
		New(tplString).
		Funcs(messageFunctions).
		Parse(tplString)
	if err != nil {
		return "", errors.Wrap(err, "parse template")
	}

	if fields == nil {
		fields = map[string]interface{}{}
	}

	fields["msg"] = m
	fields["permitTimeout"] = int64(*&config.PermitTimeout / time.Second)
	fields["username"] = m.User

	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, fields)

	return buf.String(), errors.Wrap(err, "execute template")
}

func init() {
	messageFunctions["fixUsername"] = func(username string) string { return strings.TrimLeft(username, "@") }

	messageFunctions["getArg"] = func(m *irc.Message, params ...int) (string, error) {
		msgParts := strings.Split(m.Trailing(), " ")
		if len(msgParts) < params[0]+1 {
			return "", errors.New("argument not found")
		}

		return msgParts[params[0]], nil
	}

	messageFunctions["getCounterValue"] = func(name string, _ ...string) int64 {
		return store.GetCounterValue(name)
	}

	messageFunctions["getTag"] = func(m *irc.Message, params ...string) string {
		s, _ := m.GetTag(params[0])
		return s
	}

	messageFunctions["recentGame"] = func(username string, v ...string) (string, error) {
		game, _, err := twitch.getRecentStreamInfo(username)
		if err != nil && len(v) > 0 {
			return v[0], nil
		}

		return game, err
	}
}

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

func formatMessage(tplString string, m *irc.Message, r *rule, fields map[string]interface{}) (string, error) {
	// Create anonymous functions in current context in order to access function variables
	messageFunctions := korvike.GetFunctionMap()

	messageFunctions["arg"] = func(arg int) (string, error) {
		msgParts := strings.Split(m.Trailing(), " ")
		if len(msgParts) <= arg {
			return "", errors.New("argument not found")
		}

		return msgParts[arg], nil
	}

	messageFunctions["channelCounter"] = func(name string) (string, error) {
		channel, ok := fields["channel"].(string)
		if !ok {
			return "", errors.New("channel not available")
		}

		return strings.Join([]string{channel, name}, ":"), nil
	}

	messageFunctions["counterValue"] = func(name string, _ ...string) int64 {
		return store.GetCounterValue(name)
	}

	messageFunctions["fixUsername"] = func(username string) string { return strings.TrimLeft(username, "@") }

	messageFunctions["group"] = func(idx int) (string, error) {
		fields := r.matchMessage.FindStringSubmatch(m.Trailing())
		if len(fields) <= idx {
			return "", errors.New("group not found")
		}

		return fields[idx], nil
	}

	messageFunctions["recentGame"] = func(username string, v ...string) (string, error) {
		game, _, err := twitch.getRecentStreamInfo(strings.TrimLeft(username, "#"))
		if err != nil && len(v) > 0 {
			return v[0], nil
		}

		return game, err
	}

	messageFunctions["tag"] = func(tag string) string {
		s, _ := m.GetTag(tag)
		return s
	}

	// Parse and execute template
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

	if m.Command == "PRIVMSG" && len(m.Params) > 0 {
		fields["channel"] = m.Params[0]
	}

	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, fields)

	return buf.String(), errors.Wrap(err, "execute template")
}

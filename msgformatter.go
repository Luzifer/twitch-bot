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

func formatMessage(tplString string, m *irc.Message, fields map[string]interface{}) (string, error) {
	fm := korvike.GetFunctionMap()
	fm["getArg"] = tplGetMessageArg
	fm["getCounterValue"] = tplGetCounterValue
	fm["getTag"] = tplGetTagFromMessage

	tpl, err := template.
		New(tplString).
		Funcs(fm).
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

func tplGetCounterValue(name string, _ ...string) int64 {
	return store.GetCounterValue(name)
}

func tplGetMessageArg(m *irc.Message, params ...int) (string, error) {
	msgParts := strings.Split(m.Trailing(), " ")
	if len(msgParts) < params[0]+1 {
		return "", errors.New("argument not found")
	}

	return msgParts[params[0]], nil
}

func tplGetTagFromMessage(m *irc.Message, params ...string) string {
	s, _ := m.GetTag(params[0])
	return s
}

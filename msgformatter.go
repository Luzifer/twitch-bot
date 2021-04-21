package main

import (
	"bytes"
	"text/template"
	"time"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func formatMessage(tplString string, m *irc.Message, r *rule, fields map[string]interface{}) (string, error) {
	if fields == nil {
		fields = map[string]interface{}{}
	}

	fields["msg"] = m
	fields["permitTimeout"] = int64(config.PermitTimeout / time.Second)
	fields["username"] = m.User

	if m.Command == "PRIVMSG" && len(m.Params) > 0 {
		fields["channel"] = m.Params[0]
	}

	// Parse and execute template
	tpl, err := template.
		New(tplString).
		Funcs(tplFuncs.GetFuncMap(m, r, fields)).
		Parse(tplString)
	if err != nil {
		return "", errors.Wrap(err, "parse template")
	}

	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, fields)

	return buf.String(), errors.Wrap(err, "execute template")
}

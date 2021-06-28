package main

import (
	"bytes"
	"text/template"
	"time"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

func formatMessage(tplString string, m *irc.Message, r *Rule, fields map[string]interface{}) (string, error) {
	compiledFields := map[string]interface{}{}

	if config != nil {
		configLock.RLock()
		for k, v := range config.Variables {
			compiledFields[k] = v
		}
		compiledFields["permitTimeout"] = int64(config.PermitTimeout / time.Second)
		configLock.RUnlock()
	}

	for k, v := range fields {
		compiledFields[k] = v
	}

	if m != nil {
		compiledFields["msg"] = m
		compiledFields["username"] = m.User

		if len(m.Params) > 0 {
			compiledFields["channel"] = m.Params[0]
		}
	}

	// Parse and execute template
	tpl, err := template.
		New(tplString).
		Funcs(tplFuncs.GetFuncMap(m, r, compiledFields)).
		Parse(tplString)
	if err != nil {
		return "", errors.Wrap(err, "parse template")
	}

	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, compiledFields)

	return buf.String(), errors.Wrap(err, "execute template")
}

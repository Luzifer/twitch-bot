package main

import (
	"bytes"
	"text/template"
	"time"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/go-irc/irc"
	"github.com/pkg/errors"
)

// Compile-time assertion
var _ plugins.MsgFormatter = formatMessage

func formatMessage(tplString string, m *irc.Message, r *plugins.Rule, fields *plugins.FieldCollection) (string, error) {
	compiledFields := plugins.NewFieldCollection()

	if config != nil {
		configLock.RLock()
		compiledFields.SetFromData(config.Variables)
		compiledFields.Set("permitTimeout", int64(config.PermitTimeout/time.Second))
		configLock.RUnlock()
	}

	compiledFields.SetFromData(fields.Data())

	if m != nil {
		compiledFields.Set("msg", m)
	}
	compiledFields.Set("username", plugins.DeriveUser(m, fields))
	compiledFields.Set("channel", plugins.DeriveChannel(m, fields))

	// Parse and execute template
	tpl, err := template.
		New(tplString).
		Funcs(tplFuncs.GetFuncMap(m, r, compiledFields)).
		Parse(tplString)
	if err != nil {
		return "", errors.Wrap(err, "parse template")
	}

	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, compiledFields.Data())

	return buf.String(), errors.Wrap(err, "execute template")
}

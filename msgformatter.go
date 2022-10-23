package main

import (
	"bytes"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/go-irc/irc"
	"github.com/pkg/errors"

	"github.com/Luzifer/twitch-bot/v2/plugins"
)

var (
	// Compile-time assertion
	_ plugins.MsgFormatter = formatMessage

	stripNewline = regexp.MustCompile(`(?m)\s*\n\s*`)

	formatMessageFieldSetters = []func(compiledFields *plugins.FieldCollection, m *irc.Message, fields *plugins.FieldCollection){
		formatMessageFieldChannel,
		formatMessageFieldMessage,
		formatMessageFieldUserID,
		formatMessageFieldUsername,
	}
)

func formatMessage(tplString string, m *irc.Message, r *plugins.Rule, fields *plugins.FieldCollection) (string, error) {
	compiledFields := plugins.NewFieldCollection()

	if config != nil {
		configLock.RLock()
		compiledFields.SetFromData(config.Variables)
		compiledFields.Set("permitTimeout", int64(config.PermitTimeout/time.Second))
		configLock.RUnlock()
	}

	compiledFields.SetFromData(fields.Data())

	for _, fn := range formatMessageFieldSetters {
		fn(compiledFields, m, fields)
	}

	// Template in frontend supports newlines, messages do not
	tplString = stripNewline.ReplaceAllString(tplString, " ")

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

	return strings.TrimSpace(buf.String()), errors.Wrap(err, "execute template")
}

func formatMessageFieldChannel(compiledFields *plugins.FieldCollection, m *irc.Message, fields *plugins.FieldCollection) {
	compiledFields.Set(eventFieldChannel, plugins.DeriveChannel(m, fields))
}

func formatMessageFieldMessage(compiledFields *plugins.FieldCollection, m *irc.Message, fields *plugins.FieldCollection) {
	if m == nil {
		return
	}

	compiledFields.Set("msg", m)
}

func formatMessageFieldUserID(compiledFields *plugins.FieldCollection, m *irc.Message, fields *plugins.FieldCollection) {
	if m == nil {
		return
	}

	if uid := m.Tags["user-id"]; uid != "" {
		compiledFields.Set(eventFieldUserID, uid)
	}
}

func formatMessageFieldUsername(compiledFields *plugins.FieldCollection, m *irc.Message, fields *plugins.FieldCollection) {
	compiledFields.Set("username", plugins.DeriveUser(m, fields))
}

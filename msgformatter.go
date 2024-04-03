package main

import (
	"bytes"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/go_helpers/v2/fieldcollection"
	"github.com/Luzifer/twitch-bot/v3/plugins"
)

var (
	// Compile-time assertion
	_ plugins.MsgFormatter = formatMessage

	stripNewline = regexp.MustCompile(`(?m)\s*\n\s*`)

	formatMessageFieldSetters = []func(compiledFields *fieldcollection.FieldCollection, m *irc.Message, fields *fieldcollection.FieldCollection){
		formatMessageFieldChannel,
		formatMessageFieldMessage,
		formatMessageFieldUserID,
		formatMessageFieldUsername,
	}
)

func formatMessage(tplString string, m *irc.Message, r *plugins.Rule, fields *fieldcollection.FieldCollection) (string, error) {
	compiledFields := fieldcollection.NewFieldCollection()

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

func formatMessageFieldChannel(compiledFields *fieldcollection.FieldCollection, m *irc.Message, fields *fieldcollection.FieldCollection) {
	compiledFields.Set(eventFieldChannel, plugins.DeriveChannel(m, fields))
}

func formatMessageFieldMessage(compiledFields *fieldcollection.FieldCollection, m *irc.Message, _ *fieldcollection.FieldCollection) {
	if m == nil {
		return
	}

	compiledFields.Set("msg", m)
}

func formatMessageFieldUserID(compiledFields *fieldcollection.FieldCollection, m *irc.Message, _ *fieldcollection.FieldCollection) {
	if m == nil {
		return
	}

	if uid := m.Tags["user-id"]; uid != "" {
		compiledFields.Set(eventFieldUserID, uid)
	}
}

func formatMessageFieldUsername(compiledFields *fieldcollection.FieldCollection, m *irc.Message, fields *fieldcollection.FieldCollection) {
	compiledFields.Set("username", plugins.DeriveUser(m, fields))
}

func validateTemplate(tplString string) error {
	// Template in frontend supports newlines, messages do not
	tplString = stripNewline.ReplaceAllString(tplString, " ")

	_, err := template.
		New(tplString).
		Funcs(tplFuncs.GetFuncMap(nil, nil, fieldcollection.NewFieldCollection())).
		Parse(tplString)
	return errors.Wrap(err, "parsing template")
}

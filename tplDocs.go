package main

import (
	"bytes"
	_ "embed"
	"runtime/debug"
	"sort"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

//go:embed tplDocs.tpl
var tplDocsTemplate string

func generateTplDocs() ([]byte, error) {
	tpl, err := template.New("tplDocs").Funcs(map[string]any{
		"renderExample": generateTplDocsRender,
	}).Parse(tplDocsTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "parsing tplDocs template")
	}

	sort.Slice(tplFuncs.docs, func(i, j int) bool { return tplFuncs.docs[i].Name < tplFuncs.docs[j].Name })

	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, struct {
		Funcs []plugins.TemplateFuncDocumentation
	}{
		Funcs: tplFuncs.docs,
	}); err != nil {
		return nil, errors.Wrap(err, "rendering tplDocs template")
	}

	return buf.Bytes(), nil
}

func generateTplDocsRender(e *plugins.TemplateFuncDocumentationExample) (string, error) {
	defer func() {
		if err := recover(); err != nil {
			logrus.WithError(err.(error)).Fatalf("%s", debug.Stack())
		}
	}()

	content := e.MessageContent
	if content == "" {
		content = "Hello World"
	}

	msg := &irc.Message{
		Command: "PRIVMSG",
		Params: []string{
			"#example",
			content,
		},
		Prefix: &irc.Prefix{
			Name: "exampleuser",
			User: "exampleuser",
			Host: "exampleuser.tmi.twitch.tv",
		},
		Tags: map[string]string{
			"badge-info":        "subscriber/26",
			"badges":            "moderator/1,subscriber/24",
			"color":             "#8A2BE2",
			"display-name":      "ExampleUser",
			"emotes":            "",
			"first-msg":         "0",
			"flags":             "",
			"id":                "d3167f1f-5a0c-4d78-ba68-1a6c0018d284",
			"mod":               "1",
			"returning-chatter": "0",
			"room-id":           "123456",
			"subscriber":        "1",
			"tmi-sent-ts":       "1679582970403",
			"turbo":             "0",
			"user-id":           "987654",
			"user-type":         "mod",
		},
	}

	rule := &plugins.Rule{}
	if e.MatchMessage != "" {
		rule.MatchMessage = &e.MatchMessage
	}

	return formatMessage(e.Template, msg, rule, plugins.FieldCollectionFromData(map[string]any{
		"testDuration": 5*time.Hour + 33*time.Minute + 12*time.Second,
	}))
}

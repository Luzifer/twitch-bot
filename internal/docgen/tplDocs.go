package docgen

import (
	"bytes"
	"cmp"
	_ "embed"
	"fmt"
	"runtime/debug"
	"slices"
	"text/template"
	"time"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"gopkg.in/irc.v4"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

//go:embed tplDocs.md.gotmpl
var tplDocsTemplate string

// GenerateTplDocs writes the documentation for the given template functions
func GenerateTplDocs(docs []plugins.TemplateFuncDocumentation, formatter plugins.MsgFormatter) error {
	docs = slices.Clone(docs)
	slices.SortFunc(docs, func(a, b plugins.TemplateFuncDocumentation) int { return cmp.Compare(a.Name, b.Name) })

	tpl, err := template.New("tplDocs").Funcs(map[string]any{
		"renderExample": func(example *plugins.TemplateFuncDocumentationExample) (string, error) {
			return renderTplDocsExample(example, formatter)
		},
	}).Parse(tplDocsTemplate)
	if err != nil {
		return fmt.Errorf("parsing tplDocs template: %w", err)
	}

	buf := new(bytes.Buffer)
	if err = tpl.Execute(buf, struct {
		Funcs []plugins.TemplateFuncDocumentation
	}{
		Funcs: docs,
	}); err != nil {
		return fmt.Errorf("rendering tplDocs template: %w", err)
	}

	return writeDocument(tplDocsPath, buf.Bytes())
}

func renderTplDocsExample(example *plugins.TemplateFuncDocumentationExample, formatter plugins.MsgFormatter) (out string, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("rendering template example: %v\n%s", recovered, debug.Stack())
		}
	}()

	content := example.MessageContent
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
	if example.MatchMessage != "" {
		rule.MatchMessage = &example.MatchMessage
	}

	return formatter(example.Template, msg, rule, fieldcollection.FromData(map[string]any{
		"testDuration": 5*time.Hour + 33*time.Minute + 12*time.Second,
	}))
}

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

//go:embed actorDocs.tpl
var actorDocsTemplate string

func generateActorDocs() ([]byte, error) {
	tpl, err := template.New("actorDocs").Parse(actorDocsTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing actorDocs template: %w", err)
	}

	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, struct{ Actors []plugins.ActionDocumentation }{availableActorDocs}); err != nil {
		return nil, fmt.Errorf("rendering actorDocs template: %w", err)
	}

	return buf.Bytes(), nil
}

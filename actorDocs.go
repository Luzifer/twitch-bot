package main

import (
	"bytes"
	_ "embed"
	"text/template"

	"github.com/Luzifer/twitch-bot/plugins"
	"github.com/pkg/errors"
)

//go:embed actorDocs.tpl
var actorDocsTemplate string

func generateActorDocs() ([]byte, error) {
	tpl, err := template.New("actorDocs").Parse(actorDocsTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "parsing actorDocs template")
	}

	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, struct{ Actors []plugins.ActionDocumentation }{availableActorDocs}); err != nil {
		return nil, errors.Wrap(err, "rendering actorDocs template")
	}

	return buf.Bytes(), nil
}

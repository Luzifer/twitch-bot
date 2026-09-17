package docgen

import (
	"bytes"
	"cmp"
	_ "embed"
	"fmt"
	"slices"
	"text/template"

	"github.com/Luzifer/twitch-bot/v3/plugins"
)

//go:embed actorDocs.md.gotmpl
var actorDocsTemplate string

// GenerateActorDocs writes the documentation for the given actors
func GenerateActorDocs(docs []plugins.ActionDocumentation) error {
	docs = slices.Clone(docs)
	slices.SortFunc(docs, func(a, b plugins.ActionDocumentation) int { return cmp.Compare(a.Name, b.Name) })

	tpl, err := template.New("actorDocs").Parse(actorDocsTemplate)
	if err != nil {
		return fmt.Errorf("parsing actorDocs template: %w", err)
	}

	buf := new(bytes.Buffer)
	if err = tpl.Execute(buf, struct{ Actors []plugins.ActionDocumentation }{docs}); err != nil {
		return fmt.Errorf("rendering actorDocs template: %w", err)
	}

	return writeDocument(actorDocsPath, buf.Bytes())
}

package docgen

import (
	"bytes"
	"cmp"
	_ "embed"
	"fmt"
	"slices"
	"text/template"

	"github.com/Luzifer/twitch-bot/v3/pkg/event"
)

//go:embed eventDocs.md.gotmpl
var eventDocsTemplate string

// GenerateEventDocs writes documentation for all given events.
func GenerateEventDocs(events []event.DocumentedEvent) error {
	tpl, err := template.New("eventDocs").Funcs(template.FuncMap{
		"eventFields": eventFields,
		"eventName":   eventName,
	}).Parse(eventDocsTemplate)
	if err != nil {
		return fmt.Errorf("parsing eventDocs template: %w", err)
	}

	buf := new(bytes.Buffer)
	if err = tpl.Execute(buf, struct{ Events []event.DocumentedEvent }{events}); err != nil {
		return fmt.Errorf("rendering eventDocs template: %w", err)
	}

	return writeDocument(eventDocsPath, buf.Bytes())
}

func eventFields(evt event.DocumentedEvent) []event.FieldDescription {
	fields := event.DescribeFields(evt)
	slices.SortFunc(fields, func(a, b event.FieldDescription) int { return cmp.Compare(a.Name, b.Name) })
	return fields
}

func eventName(evt event.DocumentedEvent) string { return *evt.Event() }

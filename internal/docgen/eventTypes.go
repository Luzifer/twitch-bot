package docgen

import (
	"bytes"
	_ "embed"
	"fmt"
	"reflect"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/Luzifer/twitch-bot/v3/pkg/event"
)

var (
	//go:embed eventTypes.ts.gotmpl
	eventTypesTemplate string

	reflectDurationType = reflect.TypeFor[time.Duration]()
	reflectTimeType     = reflect.TypeFor[time.Time]()
)

// GenerateEventTypes writes the TypeScript definitions for known events.
func GenerateEventTypes(events []event.DocumentedEvent) error {
	tpl, err := template.New("eventTypes").Funcs(template.FuncMap{
		"eventFields":   eventFields,
		"eventName":     eventName,
		"eventTypeName": func(evt event.DocumentedEvent) string { return eventTypeName(eventName(evt)) },
		"jsType":        jsType,
	}).Parse(eventTypesTemplate)
	if err != nil {
		return fmt.Errorf("parsing event-types template: %w", err)
	}

	buf := new(bytes.Buffer)
	if err = tpl.Execute(buf, struct{ Events []event.DocumentedEvent }{events}); err != nil {
		return fmt.Errorf("rendering event-types template: %w", err)
	}

	return writeDocument(eventTypesPath, buf.Bytes())
}

func eventTypeName(name string) string {
	var out strings.Builder
	upper := true
	for _, char := range name {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) {
			upper = true
			continue
		}
		if upper {
			char = unicode.ToUpper(char)
			upper = false
		}
		out.WriteRune(char)
	}
	return out.String()
}

func jsType(typ reflect.Type) string {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ == reflectTimeType {
		return "string"
	}
	if typ == reflectDurationType {
		return "number"
	}

	switch typ.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Float32, reflect.Float64, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "number"
	case reflect.Slice, reflect.Array:
		return "Array<" + jsType(typ.Elem()) + ">"
	case reflect.String:
		return "string"
	default:
		return "unknown"
	}
}

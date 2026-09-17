// Package event defines the interface for typed events and helpers to
// work with them
package event

import (
	"reflect"
	"slices"
	"strings"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"github.com/sirupsen/logrus"
)

type (
	// Event defines the interface for typed events
	Event interface {
		// Event returns the event type the payload represents
		Event() *string
	}

	field struct {
		Name  string
		NoLog bool
		Value any
	}
)

// ToFieldCollection converts the typed event into a FieldCollection
// for templating and handler compatibility
func ToFieldCollection(evt Event) *fieldcollection.FieldCollection {
	if evt == nil {
		return nil
	}

	extractedFields := extractFields(evt)
	if len(extractedFields) == 0 {
		return nil
	}

	fields := fieldcollection.NewFieldCollection()
	for _, field := range extractedFields {
		fields.Set(field.Name, field.Value)
	}

	return fields
}

// ToLogFields converts the typed event into logrus.Fields. Fields
// marked `nolog:"true"` will be omitted.
func ToLogFields(evt Event) logrus.Fields {
	if evt == nil {
		return nil
	}

	fields := make(logrus.Fields)
	for _, field := range extractFields(evt) {
		if field.NoLog {
			continue
		}

		fields[field.Name] = field.Value
	}

	return fields
}

func extractFields(evt Event) (flds []field) {
	var extract func(reflect.Value)

	extract = func(value reflect.Value) {
		for value.IsValid() && (value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer) {
			if value.IsNil() {
				return
			}

			value = value.Elem()
		}

		if !value.IsValid() || value.Kind() != reflect.Struct {
			return
		}

		typ := value.Type()
		for index := range typ.NumField() {
			structField := typ.Field(index)
			fieldTag := strings.Split(structField.Tag.Get("field"), ",")
			name := fieldTag[0]

			switch {
			case name == "-":
				continue

			case name != "":
				fieldValue := value.Field(index)
				if !fieldValue.CanInterface() {
					continue
				}
				if slices.Contains(fieldTag[1:], "omitzero") && isZero(fieldValue) {
					continue
				}

				flds = append(flds, field{
					Name:  name,
					NoLog: structField.Tag.Get("nolog") == "true",
					Value: unwrapFieldValue(fieldValue),
				})

			case structField.Anonymous:
				extract(value.Field(index))
			}
		}
	}

	extract(reflect.ValueOf(evt))

	return flds
}

func isZero(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return true
		}

		element := value.Elem()
		if element.Kind() == reflect.Pointer && element.IsNil() {
			return true
		}

	case reflect.Pointer:
		if value.IsNil() {
			return true
		}
	}

	if value.CanInterface() {
		if zeroer, ok := reflect.TypeAssert[interface{ IsZero() bool }](value); ok {
			return zeroer.IsZero()
		}
	}

	if value.CanAddr() && value.Addr().CanInterface() {
		if zeroer, ok := reflect.TypeAssert[interface{ IsZero() bool }](value.Addr()); ok {
			return zeroer.IsZero()
		}
	}

	return value.IsZero()
}

func unwrapFieldValue(value reflect.Value) any {
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		value = value.Elem()
	}

	return value.Interface()
}

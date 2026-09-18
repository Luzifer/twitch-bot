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

	// FieldDescription describes a field exposed by an event.
	FieldDescription struct {
		Description string
		Name        string
		Optional    bool
		Type        reflect.Type
	}
)

// DescribeFields returns the fields exposed by evt based on its struct tags.
func DescribeFields(evt Event) (flds []FieldDescription) {
	walkFields(reflect.TypeOf(evt), func(structField reflect.StructField, name string, optional bool) {
		flds = append(flds, FieldDescription{
			Description: structField.Tag.Get("description"),
			Name:        name,
			Optional:    optional,
			Type:        structField.Type,
		})
	})

	return flds
}

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
	value := reflect.ValueOf(evt)
	for value.IsValid() && (value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer) {
		if value.IsNil() {
			return nil
		}

		value = value.Elem()
	}

	if !value.IsValid() || value.Kind() != reflect.Struct {
		return nil
	}

	walkFields(reflect.TypeOf(evt), func(structField reflect.StructField, name string, optional bool) {
		fieldValue := fieldByIndex(value, structField.Index)
		if fieldValue.IsValid() {
			if !fieldValue.CanInterface() {
				return
			}
			if optional && isZero(fieldValue) {
				return
			}

			flds = append(flds, field{
				Name:  name,
				NoLog: structField.Tag.Get("nolog") == "true",
				Value: unwrapFieldValue(fieldValue),
			})
		}
	})

	return flds
}

func fieldByIndex(value reflect.Value, index []int) reflect.Value {
	for _, fieldIndex := range index {
		for value.IsValid() && value.Kind() == reflect.Pointer {
			if value.IsNil() {
				return reflect.Value{}
			}
			value = value.Elem()
		}
		if !value.IsValid() || value.Kind() != reflect.Struct {
			return reflect.Value{}
		}
		value = value.Field(fieldIndex)
	}

	return value
}

func walkFields(typ reflect.Type, fn func(reflect.StructField, string, bool)) {
	for typ != nil && (typ.Kind() == reflect.Interface || typ.Kind() == reflect.Pointer) {
		typ = typ.Elem()
	}
	if typ == nil || typ.Kind() != reflect.Struct {
		return
	}

	for structField := range typ.Fields() {
		fieldTag := strings.Split(structField.Tag.Get("field"), ",")
		name := fieldTag[0]

		switch {
		case name == "-":
			continue
		case name != "":
			fn(structField, name, slices.Contains(fieldTag[1:], "omitzero"))
		case structField.Anonymous:
			walkFields(structField.Type, func(embeddedField reflect.StructField, embeddedName string, optional bool) {
				embeddedField.Index = append(append([]int(nil), structField.Index...), embeddedField.Index...)
				fn(embeddedField, embeddedName, optional)
			})
		}
	}
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

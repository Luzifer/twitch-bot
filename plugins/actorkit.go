package plugins

import (
	"reflect"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"github.com/pkg/errors"
)

type (
	// ActorKit contains some common validation functions to be used
	// when implementing actors
	ActorKit struct{}
)

// ValidateRequireNonEmpty checks whether the fields are gettable
// (not returning ErrValueNotSet) and does not contain zero value
// recognized by reflect (to just check whether the field is set
// but allow zero values use HasAll on the FieldCollection)
func (ActorKit) ValidateRequireNonEmpty(attrs *fieldcollection.FieldCollection, fields ...string) error {
	for _, field := range fields {
		v, err := attrs.Get(field)
		if err != nil {
			return errors.Wrapf(err, "getting field %s", field)
		}

		if reflect.ValueOf(v).IsZero() {
			return errors.Errorf("field %s has zero-value", field)
		}
	}

	return nil
}

// ValidateRequireValidTemplate checks whether fields are gettable
// as strings and do have a template which validates (this does not
// check for empty strings as an empty template is indeed valid)
func (ActorKit) ValidateRequireValidTemplate(tplValidator TemplateValidatorFunc, attrs *fieldcollection.FieldCollection, fields ...string) error {
	for _, field := range fields {
		v, err := attrs.String(field)
		if err != nil {
			return errors.Wrapf(err, "getting string field %s", field)
		}

		if err = tplValidator(v); err != nil {
			return errors.Wrapf(err, "validaging template field %s", field)
		}
	}

	return nil
}

// ValidateRequireValidTemplateIfSet checks whether the field is
// either not set or a valid template (this does not
// check for empty strings as an empty template is indeed valid)
func (ActorKit) ValidateRequireValidTemplateIfSet(tplValidator TemplateValidatorFunc, attrs *fieldcollection.FieldCollection, fields ...string) error {
	for _, field := range fields {
		v, err := attrs.String(field)
		if err != nil {
			if errors.Is(err, fieldcollection.ErrValueNotSet) {
				continue
			}
			return errors.Wrapf(err, "getting string field %s", field)
		}

		if err = tplValidator(v); err != nil {
			return errors.Wrapf(err, "validaging template field %s", field)
		}
	}

	return nil
}

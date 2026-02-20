package helpers

import (
	"fmt"

	"github.com/Luzifer/go_helpers/fieldcollection"
)

// SchemaValidateTemplateField contains a ValidateOpt for the
// fieldcollection schema validator to validate template fields
func SchemaValidateTemplateField(tplValidator func(string) error, fields ...string) fieldcollection.ValidateOpt {
	return func(f, _ *fieldcollection.FieldCollection) (err error) {
		for _, field := range fields {
			if err = tplValidator(f.MustString(field, Ptr(""))); err != nil {
				return fmt.Errorf("validating %s: %w", field, err)
			}
		}

		return nil
	}
}

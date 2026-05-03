package plugins

import (
	"strings"
	"testing"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"github.com/stretchr/testify/require"
)

func TestValidateRequireNonEmpty(t *testing.T) {
	attrs := fieldcollection.FieldCollectionFromData(map[string]any{
		"str":   "",
		"str_v": "valid",
		"int":   0,
		"int_v": 1,
	})

	for _, field := range []string{"int", "str"} {
		errUnset := ActorKit{}.ValidateRequireNonEmpty(attrs, strings.Join([]string{field, "unset"}, "_"))
		errInval := ActorKit{}.ValidateRequireNonEmpty(attrs, field)
		errValid := ActorKit{}.ValidateRequireNonEmpty(attrs, strings.Join([]string{field, "v"}, "_"))

		require.Error(t, errUnset)
		require.Error(t, errInval)
		require.NoError(t, errValid)
	}
}

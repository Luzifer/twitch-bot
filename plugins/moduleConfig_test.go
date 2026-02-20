package plugins

import (
	"testing"

	"github.com/Luzifer/go_helpers/fieldcollection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModuleConfigGet(t *testing.T) {
	strPtrEmpty := func(v string) *string { return &v }("")
	m := ModuleConfig{
		"test": map[string]*fieldcollection.FieldCollection{
			DefaultConfigName: fieldcollection.FieldCollectionFromData(map[string]any{
				"setindefault": DefaultConfigName,
				"setinboth":    DefaultConfigName,
			}),
			"test": fieldcollection.FieldCollectionFromData(map[string]any{
				"setinchannel": "channel",
				"setinboth":    "channel",
			}),
		},
	}

	fields := m.GetChannelConfig("module_does_not_exist", "test")
	require.NotNil(t, fields, "must always return a valid FieldCollection")
	assert.Len(t, fields.Data(), 0)

	fields = m.GetChannelConfig("test", "test")
	assert.Equal(t, DefaultConfigName, fields.MustString("setindefault", strPtrEmpty))
	assert.Equal(t, "channel", fields.MustString("setinchannel", strPtrEmpty))
	assert.Equal(t, "channel", fields.MustString("setinboth", strPtrEmpty))

	fields = m.GetChannelConfig("test", "channel_not_configured")
	assert.Equal(t, DefaultConfigName, fields.MustString("setindefault", strPtrEmpty))
	assert.Equal(t, "", fields.MustString("setinchannel", strPtrEmpty))
	assert.Equal(t, DefaultConfigName, fields.MustString("setinboth", strPtrEmpty))
}

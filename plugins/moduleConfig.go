package plugins

import (
	"strings"

	"github.com/Luzifer/go_helpers/v2/fieldcollection"
)

// DefaultConfigName is the name the default configuration must have
// when defined
const DefaultConfigName = "default"

type (
	// ModuleConfig represents a mapping of configurations per channel
	// and module
	ModuleConfig map[string]map[string]*fieldcollection.FieldCollection
)

// GetChannelConfig reads the channel specific configuration for the
// given module. This is created by taking an empty FieldCollection,
// merging in the default configuration and finally overwriting all
// existing channel configurations.
func (m ModuleConfig) GetChannelConfig(module, channel string) *fieldcollection.FieldCollection {
	channel = strings.TrimLeft(channel, "#@")
	composed := fieldcollection.NewFieldCollection()

	for _, i := range []string{DefaultConfigName, channel} {
		f := m[module][i]
		if f == nil {
			// That config does not exist, don't apply
			continue
		}

		for k, v := range f.Data() {
			// Overwrite all keys defined in this config
			composed.Set(k, v)
		}
	}

	return composed
}

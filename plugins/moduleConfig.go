package plugins

import "strings"

// DefaultConfigName is the name the default configuration must have
// when defined
const DefaultConfigName = "default"

type (
	// ModuleConfig represents a mapping of configurations per channel
	// and module
	ModuleConfig map[string]map[string]*FieldCollection
)

// GetChannelConfig reads the channel specific configuration for the
// given module. This is created by taking an empty FieldCollection,
// merging in the default configuration and finally overwriting all
// existing channel configurations.
func (m ModuleConfig) GetChannelConfig(module, channel string) *FieldCollection {
	channel = strings.TrimLeft(channel, "#@")
	composed := NewFieldCollection()

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

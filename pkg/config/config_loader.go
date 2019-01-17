package config

import (
	yaml "gopkg.in/yaml.v2"
)

// SourcesProvider is an interface which provides a slice of configuration source loader functions as strategies
type SourcesProvider interface {
	Sources() []Source
}

// Source is a function type representing strategy for loading configuration file into []byte
type Source func() ([]byte, error)

// PluginConfiguration holds common configuration for all the plugins
type PluginConfiguration struct {
	PluginName  string
	LocationURL string
}

// Load loads configuration of the plugin based on strategies defined by SourcesProvider
// It ignores errors returned by providers and only propagates the one occurred while unmarshalling
func Load(target interface{}, loader SourcesProvider) error {
	var source []byte
	for _, load := range loader.Sources() {
		loaded, err := load()
		if err == nil {
			source = loaded
			break
		}
	}
	return yaml.Unmarshal(source, target)
}

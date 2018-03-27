package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// SourcesProvider is an interface which provides a slice of configuration source loader functions as strategies
type SourcesProvider interface {
	Sources() []Source
}

// Source is a function type representing strategy for loading configuration file into []byte
type Source func() ([]byte, error)

// PluginConfiguration holds common configuration for all the plugins
type PluginConfiguration struct {
	LocationURL string
	PluginHint  string `yaml:"plugin_hint,omitempty"`
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

// LocalLoadableConfig holds absolute path to a local config file (located in ike-prow-plugins project) to be loaded
type LocalLoadableConfig struct {
	AbsFilePath string
}

// Sources loads local config file that is located in ike-prow-plugins project structure
func (i *LocalLoadableConfig) Sources() []Source {
	return []Source{func() ([]byte, error) {
		file, err := ioutil.ReadFile(i.AbsFilePath)
		if err != nil {
			return nil, err
		}
		return file, nil
	}}
}

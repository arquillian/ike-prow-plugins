package config

import (
	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	"gopkg.in/yaml.v2"
)

// PluginConfiguration holds common configuration for all the plugins
type PluginConfiguration struct {
	LocationURL string
	PluginHint  string `yaml:"plugin_hint,omitempty"`
}

// Load loads configuration of the plugin stored in the YAML file named after the plugin name
// It looks it up based on the scm.RepositoryChange hash information and unmarshals content into
// passed target interface
func Load(target interface{}, sources ...func() ([]byte, error)) error {
	var source []byte
	for _, load := range sources {
		loaded, err := load()
		if err == nil {
			source = loaded
			break
		}
	}
	return yaml.Unmarshal(source, target)
}

// TODO might need a new home
// nolint
func GitHubConfigLoader(baseConfig *PluginConfiguration, filePath string, change scm.RepositoryChange) func() ([]byte, error) {
	rawFileService := github.RawFileService{
		Change: change,
	}

	return func() ([]byte, error) {
		configURL := rawFileService.GetRawFileURL(filePath)
		downloadedConfig, err := utils.GetFileFromURL(configURL)

		if err != nil {
			return nil, err
		}
		baseConfig.LocationURL = configURL
		return downloadedConfig, nil
	}
}

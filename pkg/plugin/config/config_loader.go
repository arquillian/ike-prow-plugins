package config

import (
	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"gopkg.in/yaml.v2"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	"errors"
)

// Configuration is an interface representing a config that contains location URL when it is downloaded
type Configuration interface {
	setLocationURL(locationURL string)
}

// PluginConfiguration is a very basic implementation of Configuration interface
type PluginConfiguration struct {
	LocationURL string
}

// PluginConfigLoader is a struct representing plugin configuration loading service
type PluginConfigLoader struct {
	pluginName     string
	rawFileService github.RawFileService
}

// NewPluginConfigLoader creates PluginConfigLoader with the given pluginName and a github.RawFileService with the given change
func NewPluginConfigLoader(pluginName string, change scm.RepositoryChange) *PluginConfigLoader {
	return &PluginConfigLoader{
		pluginName: pluginName,
		rawFileService: github.RawFileService{
			Change: change,
		},
	}
}

func (config *PluginConfiguration) setLocationURL(locationURL string) {
	config.LocationURL = locationURL
}

// Load loads configuration of the plugin stored in the YAML file named after the plugin name
// It looks it up based on the scm.RepositoryChange hash information and unmarshals content into
// passed target interface
func (loader *PluginConfigLoader) Load(config Configuration) error {

	var downloadedConfig []byte
	configURL := loader.rawFileService.GetRawFileURL(fmt.Sprintf("%s.yml", loader.pluginName))
	downloadedConfig, ymlErr := utils.GetFileFromURL(configURL)

	if ymlErr != nil {
		configURL = loader.rawFileService.GetRawFileURL(fmt.Sprintf("%s.yaml", loader.pluginName))
		var yamlErr error
		downloadedConfig, yamlErr = utils.GetFileFromURL(configURL)

		if yamlErr != nil {
			return errors.New(fmt.Sprintf("yml error: %s\nyaml error: %s", ymlErr.Error(), yamlErr.Error()))
		}
	}

	config.setLocationURL(configURL)
	return yaml.Unmarshal(downloadedConfig, config)
}

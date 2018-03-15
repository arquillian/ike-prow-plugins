package config

import (
	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"gopkg.in/yaml.v2"
)

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

// Load loads configuration of the plugin stored in the YAML file named after the plugin name
// It looks it up based on the scm.RepositoryChange hash information and unmarshals content into
// passed target interface
func (loader *PluginConfigLoader) Load(target interface{}) error {
	configuration, err := loader.rawFileService.GetRawFile(loader.createConfigFilePath())
	if err != nil {
		return err
	}
	return yaml.Unmarshal(configuration, target)
}

// CreateConfigFileURL creates a url to the configuration file
func (loader *PluginConfigLoader) CreateConfigFileURL() string {
	return loader.rawFileService.GetRawFileURL(loader.createConfigFilePath())
}

func (loader *PluginConfigLoader) createConfigFilePath() string {
	return fmt.Sprintf("%s.yml", loader.pluginName)
}

package testkeeper

import (
	"github.com/arquillian/ike-prow-plugins/pkg/config"
	ghservice "github.com/arquillian/ike-prow-plugins/pkg/github/service"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
)

// PluginConfiguration defines inclusion and exclusion patterns set of files will be matched against
// It's unmarshaled from test-keeper.yml configuration file
type PluginConfiguration struct {
	config.PluginConfiguration `yaml:",inline,omitempty"`
	Inclusions                 []string `yaml:"test_patterns,omitempty"`
	Exclusions                 []string `yaml:"skip_validation_for,omitempty"`
	Combine                    bool     `yaml:"combine_defaults,omitempty"`
}

// LoadConfiguration loads a PluginConfiguration for the given change
func LoadConfiguration(logger log.Logger, change scm.RepositoryChange) *PluginConfiguration {

	configuration := PluginConfiguration{Combine: true}
	loadableConfig := &ghservice.LoadableConfig{PluginName: ProwPluginName, Change: change, BaseConfig: &configuration.PluginConfiguration}

	err := config.Load(&configuration, loadableConfig)

	if err != nil {
		logger.Errorf("Config file was not loaded. Cause: %s", err)
		return &configuration
	}

	return &configuration
}

package testkeeper

import (
	"github.com/arquillian/ike-prow-plugins/pkg/config"
	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
)

// PluginConfiguration defines inclusion and exclusion patterns set of files will be matched against
// It's unmarshaled from test-keeper.yml configuration file
type PluginConfiguration struct {
	config.PluginConfiguration `yaml:",inline"`
	Inclusions                 []string `yaml:"test_patterns,omitempty"`
	Exclusions                 []string `yaml:"skip_validation_for,omitempty"`
	Combine                    bool     `yaml:"combine_defaults,omitempty"`
}

// LoadConfiguration loads a PluginConfiguration for the given change
func LoadConfiguration(log log.Logger, change scm.RepositoryChange) PluginConfiguration {

	configuration := PluginConfiguration{Combine: true}
	loadableConfig := &github.LoadableConfig{PluginName: ProwPluginName, Change: change, BaseConfig: &configuration.PluginConfiguration}

	err := config.Load(&configuration, loadableConfig)

	if err != nil {
		log.Warnf("Config file was not loaded. Cause: %s", err)
		return configuration
	}

	return configuration
}

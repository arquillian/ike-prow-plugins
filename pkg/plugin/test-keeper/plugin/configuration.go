package plugin

import (
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/config"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
)

// TestKeeperConfiguration defines inclusion and exclusion patterns set of files will be matched against
// It's unmarshaled from test-keeper.yml configuration file
type TestKeeperConfiguration struct {
	config.PluginConfiguration
	Inclusion   string `yaml:"test_pattern,omitempty"`
	Exclusion   string `yaml:"skip_validation_for,omitempty"`
	Combine     bool   `yaml:"combine_defaults,omitempty"`
	PluginHint  string `yaml:"plugin_hint,omitempty"`
}

// LoadTestKeeperConfig loads a TestKeeperConfiguration for the given change
func LoadTestKeeperConfig(log log.Logger, change scm.RepositoryChange) TestKeeperConfiguration {
	configLoader := config.NewPluginConfigLoader(ProwPluginName, change)

	configuration := TestKeeperConfiguration{Combine: true}
	err := configLoader.Load(&configuration)
	if err != nil {
		log.Warnf("Config file was not loaded. Cause: %s", err)
		return configuration
	}
	return configuration
}

package plugin

import (
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/config"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/sirupsen/logrus"
)

// TestKeeperConfiguration defines inclusion and exclusion patterns set of files will be matched against
// It's unmarshaled from test-keeper.yml configuration file
type TestKeeperConfiguration struct {
	Inclusion  string `yaml:"test_pattern,omitempty"`
	Exclusion  string `yaml:"skip_validation_for,omitempty"`
	Combine    bool   `yaml:"combine_defaults,omitempty"`
	PluginHint string `yaml:"plugin_hint,omitempty"`
}

// LoadTestKeeperConfig loads a TestKeeperConfiguration for the given change
func LoadTestKeeperConfig(log *logrus.Entry, change scm.RepositoryChange) (urlIfExists string, conf TestKeeperConfiguration) {
	configLoader := config.NewPluginConfigLoader(ProwPluginName, change)

	configuration := TestKeeperConfiguration{Combine: true}
	exists, err := configLoader.Load(&configuration)
	if err != nil {
		log.Warnf("Config file was not loaded. Cause: %", err)
	}
	if exists {
		return configLoader.CreateConfigFileURL(), configuration
	}
	return "", configuration
}

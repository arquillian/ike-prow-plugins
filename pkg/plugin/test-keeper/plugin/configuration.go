package plugin

import (
	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/config"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
)

// TestKeeperConfiguration defines inclusion and exclusion patterns set of files will be matched against
// It's unmarshaled from test-keeper.yml configuration file
type TestKeeperConfiguration struct {
	BaseConfig config.PluginConfiguration `yaml:",omitempty,inline"`
	Inclusion  string                     `yaml:"test_pattern,omitempty"`
	Exclusion  string                     `yaml:"skip_validation_for,omitempty"`
	Combine    bool                       `yaml:"combine_defaults,omitempty"`
}

// LoadTestKeeperConfig loads a TestKeeperConfiguration for the given change
func LoadTestKeeperConfig(log log.Logger, change scm.RepositoryChange) TestKeeperConfiguration {

	configuration := TestKeeperConfiguration{Combine: true, BaseConfig: config.PluginConfiguration{}}

	err := config.Load(&configuration,
		config.GitHubConfigLoader(&configuration.BaseConfig, fmt.Sprintf("%s.yml", "test-keeper"), change),
		config.GitHubConfigLoader(&configuration.BaseConfig, fmt.Sprintf("%s.yaml", "test-keeper"), change))

	if err != nil {
		log.Warnf("Config file was not loaded. Cause: %s", err)
		return configuration
	}

	return configuration
}

package prsanitizer

import (
	"github.com/arquillian/ike-prow-plugins/pkg/config"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
)

// PluginConfiguration defines prefix patterns set of PR titles will be matched against
// It's unmarshaled from work-in-progress.yml configuration file
type PluginConfiguration struct {
	config.PluginConfiguration `yaml:",inline,omitempty"`
	TypePrefix                     []string `yaml:"type_prefixes,omitempty"`
	Combine                    bool     `yaml:"combine_defaults,omitempty"`
}

// LoadConfiguration loads a PluginConfiguration for the given change
func LoadConfiguration(log log.Logger, change scm.RepositoryChange) PluginConfiguration {

	configuration := PluginConfiguration{Combine: true}
	loadableConfig := &ghservice.LoadableConfig{PluginName: ProwPluginName, Change: change, BaseConfig: &configuration.PluginConfiguration}

	err := config.Load(&configuration, loadableConfig)

	if err != nil {
		log.Errorf("Config file was not loaded. Cause: %s", err)
		return configuration
	}

	return configuration
}

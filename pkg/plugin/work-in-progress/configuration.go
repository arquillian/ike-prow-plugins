package wip

import (
	"github.com/arquillian/ike-prow-plugins/pkg/config"
	ghservice "github.com/arquillian/ike-prow-plugins/pkg/github/service"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
)

// PluginConfiguration defines prefix patterns set of PR titles will be matched against
// It's unmarshaled from work-in-progress.yml configuration file
type PluginConfiguration struct {
	config.PluginConfiguration `yaml:",inline,omitempty"`
	Prefix                     []string `yaml:"title_prefixes,omitempty"`
	Label                      string   `yaml:"gh_label,omitempty"`
	Combine                    bool     `yaml:"combine_defaults,omitempty"`
}

// DefaultLabel is the GitHub label name set in absence of any configured label name
const DefaultLabel = "work-in-progress"

// LoadConfiguration loads a PluginConfiguration for the given change
func LoadConfiguration(logger log.Logger, change scm.RepositoryChange) PluginConfiguration {

	configuration := PluginConfiguration{Combine: true, Label: DefaultLabel}
	loadableConfig := &ghservice.LoadableConfig{PluginName: ProwPluginName, Change: change, BaseConfig: &configuration.PluginConfiguration}

	err := config.Load(&configuration, loadableConfig)

	if err != nil {
		logger.Errorf("Config file was not loaded. Cause: %s", err)
		return configuration
	}

	return configuration
}

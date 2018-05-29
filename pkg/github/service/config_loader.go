package ghservice

import (
	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/config"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
)

const githubBaseURL = "https://github.com/"

// ConfigHome is a directory to keep prow configuration files
const ConfigHome = ".ike-prow/"

// LoadableConfig holds information about the plugin name, repository change and pointer to base config
type LoadableConfig struct {
	PluginName string
	Change     scm.RepositoryChange
	BaseConfig *config.PluginConfiguration
}

// Sources provides default loading strategies for a plugin looking it up in the .ike-prow directory of the repository for a given
// revision. Two files are expected to be found there plugin-name.yml or plugin-name.yaml (in that order)
func (l *LoadableConfig) Sources() []config.Source {
	return []config.Source{
		l.loadFromRawFile(ConfigHome + "%s.yml"),
		l.loadFromRawFile(ConfigHome + "%s.yaml"),
	}
}

func (l *LoadableConfig) loadFromRawFile(pathTemplate string) config.Source {

	filePath := fmt.Sprintf(pathTemplate, l.PluginName)

	rawFileService := RawFileService{
		Change: l.Change,
	}

	return func() ([]byte, error) {
		configURL := rawFileService.GetRawFileURL(filePath)
		downloadedConfig, err := utils.GetFileFromURL(configURL)
		l.BaseConfig.PluginName = l.PluginName

		if err != nil {
			return nil, err
		}
		l.BaseConfig.LocationURL = githubBaseURL + rawFileService.GetRelativePath(filePath)

		return downloadedConfig, nil
	}
}

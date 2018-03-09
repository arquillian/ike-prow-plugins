package config

import (
	"github.com/arquillian/ike-prow-plugins/scm"
	"fmt"
	"net/http"
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

// PluginConfigLoader is a struct representing plugin configuration loading service
type PluginConfigLoader struct {
	PluginName string
	Change     scm.RepositoryChange
}

// Load loads configuration of the plugin stored in the YAML file named after the plugin name
// It looks it up based on the scm.RepositoryChange hash information and unmarshals content into
// passed target interface
func (loader *PluginConfigLoader) Load(target interface{}) error {
	path := fmt.Sprintf("%s.yml", loader.PluginName)

	configuration, err := loader.getRawFile(loader.Change.Owner, loader.Change.RepoName, loader.Change.Hash, path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(configuration, target)
}

func (loader *PluginConfigLoader) getRawFile(owner, repo, sha, path string) ([]byte, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repo, sha, path)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

package plugin

import (
	"fmt"
	"net/url"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
)

const (
	beginning = "There wasn't found any test file that would be changed in this PR. Please add a test case to this PR.\n" +
		"If you are an admin and you are sure that no test is needed then you can use command `" + SkipComment + "` " +
		"as a comment to make the status green.\n"

	noConfig = "For more information about the plugin and how to configure it, go to the " +
		"[documentation](http://arquillian.org/ike-prow-plugins/#_test_keeper_plugin)."

	withConfig = "To check plugin configuration that is set for your repository, please see the [configuration file](%s)."

	notFoundFileSuffix = "\nIn the configuration file, there is set a path to a file containing a custom comment message, " +
		"but the plugin wasn't able to retrieve it from the location %s. Check if it is either a valid URL or a relative " +
		"path to an existing file in this repository."
)

// CreateCommentMessage creates a comment message for the test-keeper plugin. If the comment message is set in config then it takes that one, the default otherwise.
func CreateCommentMessage(urlToConfig string, configuration TestKeeperConfiguration, change scm.RepositoryChange) string {
	if urlToConfig == "" {
		return beginning + noConfig
	}
	if configuration.PluginHint != "" {
		return getMsgFromFile(urlToConfig, configuration, change)
	}
	return getMsgWithConfigRef(urlToConfig)

}

func getMsgFromFile(urlToConfig string, configuration TestKeeperConfiguration, change scm.RepositoryChange) string {
	_, err := url.ParseRequestURI(configuration.PluginHint)

	var content []byte
	var msgFileURL string

	if err == nil {
		msgFileURL = configuration.PluginHint
		content, err = utils.GetFileFromURL(configuration.PluginHint)
	} else {
		ghFileService := github.RawFileService{Change: change}
		msgFileURL = ghFileService.GetRawFileURL(configuration.PluginHint)
		content, err = ghFileService.GetRawFile(configuration.PluginHint)
	}

	if err != nil {
		return getMsgWithConfigRef(urlToConfig) + fmt.Sprintf(notFoundFileSuffix, msgFileURL)
	}

	return string(content)
}

func getMsgWithConfigRef(urlToConfig string) string {
	return beginning + fmt.Sprintf(withConfig, urlToConfig)
}

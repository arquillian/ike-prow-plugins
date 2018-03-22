package plugin

import (
	"fmt"
	"net/url"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
)

const (
	beginning = "Thanks for this contribution! It seems that there are no tests that would be added or changed in this PR. " +
		"Automated tests give us confidence in shipping reliable software. Please add some as part of this change.\n" +
		"If you are an admin and you are sure that no test is needed then you can use command `" + SkipComment + "` " +
		"as a comment to make the status green.\n"

	noConfig = "For more information about the plugin and how to configure it, go to the " +
		"[documentation](http://arquillian.org/ike-prow-plugins/#_test_keeper_plugin)."

	withConfig = "Your plugin configuration is stored in the [file](%s)."

	notFoundFileSuffix = "\nIn the configuration file you pointed to the custom comment file, " +
		"but the plugin wasn't able to retrieve it from the defined location (%s). Make sure it is either a valid URL or a valid " +
		"path to an existing file in this repository."
)

// CreateCommentMessage creates a comment message for the test-keeper plugin. If the comment message is set in config then it takes that one, the default otherwise.
func CreateCommentMessage(configuration TestKeeperConfiguration, change scm.RepositoryChange) string {
	if configuration.BaseConfig.LocationURL == "" {
		return beginning + noConfig
	}
	if configuration.BaseConfig.PluginHint != "" {
		return getMsgFromFile(configuration, change)
	}
	return getMsgWithConfigRef(configuration.BaseConfig.LocationURL)

}

func getMsgFromFile(configuration TestKeeperConfiguration, change scm.RepositoryChange) string {
	_, err := url.ParseRequestURI(configuration.BaseConfig.PluginHint)

	var content []byte
	var msgFileURL string

	if err == nil {
		msgFileURL = configuration.BaseConfig.PluginHint
	} else {
		ghFileService := github.RawFileService{Change: change}
		msgFileURL = ghFileService.GetRawFileURL(configuration.BaseConfig.PluginHint)
	}
	content, err = utils.GetFileFromURL(msgFileURL)

	if err != nil {
		return getMsgWithConfigRef(configuration.BaseConfig.LocationURL) + fmt.Sprintf(notFoundFileSuffix, msgFileURL)
	}

	return string(content)
}

func getMsgWithConfigRef(urlToConfig string) string {
	return beginning + fmt.Sprintf(withConfig, urlToConfig)
}

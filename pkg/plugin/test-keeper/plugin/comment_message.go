package plugin

import (
	"fmt"
	"net/url"
	"regexp"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/http"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
)

const (
	paragraph = "\n\n"

	beginning = "It appears that no tests have been added or updated in this PR." +
		paragraph +
		"Automated tests give us confidence in shipping reliable software. Please add some as part of this change." +
		paragraph +
		"If you are an admin and you are sure that no test is needed then you can use the command `" + SkipComment + "` " +
		"as a comment to make the status green.\n"

	noConfig = "For more information please head over to official " +
		"[documentation](http://arquillian.org/ike-prow-plugins/#_test_keeper_plugin). You can find there how to " +
		"configure the plugin - for example exclude certain file types so if PR contains only them it won't be checked."

	withConfig = "Your plugin configuration is stored in the [file](%s)."

	notFoundFileSuffix = "In the configuration file you pointed to the custom comment file, " +
		"but the plugin wasn't able to retrieve it from the defined location (%s). Make sure it is either a valid URL or a " +
		"path to an existing file in this repository."

	sadIke = `<img align="left" src="https://cdn.rawgit.com/bartoszmajsak/ike-prow-plugins/2025328b70bd1879520411b3cacadee61a49641a/docs/images/arquillian_ui_failure_128px.png">`

	fileRegex = `(?m)^(?:https?\:)?([a-z_\-\s0-9\.\/]+)+\.(txt|md|doc|docx|adoc)$`
)

// CreateCommentMessage creates a comment message for the test-keeper plugin. If the comment message is set in config then it takes that one, the default otherwise.
func CreateCommentMessage(configuration TestKeeperConfiguration, change scm.RepositoryChange) string {
	var msg string
	if configuration.LocationURL == "" {
		msg = beginning + paragraph + noConfig
	} else if configuration.PluginHint != "" {
		msg = getMsgFromConfigHint(configuration, change)
	} else {
		msg = getMsgWithConfigRef(configuration.LocationURL)
	}
	return sadIke + paragraph + msg
}

func getMsgFromConfigHint(configuration TestKeeperConfiguration, change scm.RepositoryChange) string {
	isFilePath, _ := regexp.MatchString(fileRegex, configuration.PluginHint)
	if isFilePath {
		return getMsgFromFile(configuration, change)
	}
	return configuration.PluginHint
}

func getMsgFromFile(configuration TestKeeperConfiguration, change scm.RepositoryChange) string {
	_, err := url.ParseRequestURI(configuration.PluginHint)

	var content []byte
	var msgFileURL string

	if err == nil {
		msgFileURL = configuration.PluginHint
	} else {
		ghFileService := github.RawFileService{Change: change}
		msgFileURL = ghFileService.GetRawFileURL(configuration.PluginHint)
	}
	content, err = http.GetFileFromURL(msgFileURL)

	if err != nil {
		return getMsgWithConfigRef(configuration.LocationURL) + paragraph + fmt.Sprintf(notFoundFileSuffix, msgFileURL)
	}

	return string(content)
}

func getMsgWithConfigRef(urlToConfig string) string {
	return beginning + paragraph + fmt.Sprintf(withConfig, urlToConfig)
}

package comment

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"text/template"

	assets "github.com/arquillian/ike-prow-plugins/pkg/assets/generated"
	"github.com/arquillian/ike-prow-plugins/pkg/config"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
)

// Message struct for the plugin comment service
type Message struct {
	Thumbnail      string
	Description    string
	ConfigFile     string
	Documentation  string
	MessageFileURL string
}

var commentMessage Message

// LoadMessageTemplate loads a comment message for the plugins from the template files. If the comment message is set in config then it takes that one, the default otherwise.
func LoadMessageTemplate(message Message, configuration config.PluginConfiguration, change scm.RepositoryChange) string {
	var msg string
	commentMessage = message
	if configuration.PluginHint != "" {
		msg = getMsgFromConfigHint(configuration, change)
	} else if content := defaultFileContent(configuration, change); content != "" {
		msg = content
	} else if configuration.LocationURL == "" {
		msg = loadMessageTemplate("message-with-no-config.txt")
	} else {
		msg = loadMessageTemplate("message-with-config.txt")
	}
	return getMsgFromTemplate(msg)
}

func getMsgFromTemplate(msg string) string {
	var tpl bytes.Buffer
	msgTemplate := template.Must(template.New("message").Parse(msg))
	err := msgTemplate.Execute(&tpl, commentMessage)
	if err != nil {
		return ""
	}
	return tpl.String()
}

func getMsgFromConfigHint(configuration config.PluginConfiguration, change scm.RepositoryChange) string {
	fileRegex := "(?mi)" + configuration.PluginName + "_hint.md$"

	isFilePath, err := regexp.MatchString(fileRegex, configuration.PluginHint)
	if isFilePath && err == nil {
		return getMsgFromFile(configuration, change)
	}
	return configuration.PluginHint
}

func defaultFileContent(configuration config.PluginConfiguration, change scm.RepositoryChange) string {
	pluginHintPath := fmt.Sprintf("%s%s_hint.md", ghservice.ConfigHome, configuration.PluginName)
	ghFileService := ghservice.RawFileService{Change: change}
	hintURL := ghFileService.GetRawFileURL(pluginHintPath)

	content, e := utils.GetFileFromURL(hintURL)
	if e != nil {
		return ""
	}
	return string(content)
}

func getMsgFromFile(configuration config.PluginConfiguration, change scm.RepositoryChange) string {
	_, err := url.ParseRequestURI(configuration.PluginHint)

	var content []byte
	var msgFileURL string

	if err == nil {
		msgFileURL = configuration.PluginHint
	} else {
		ghFileService := ghservice.RawFileService{Change: change}
		msgFileURL = ghFileService.GetRawFileURL(configuration.PluginHint)
	}
	content, err = utils.GetFileFromURL(msgFileURL)

	if err != nil {
		commentMessage.MessageFileURL = msgFileURL
		return loadMessageTemplate("message-with-file-not-found.txt")
	}

	return string(content)
}

func loadMessageTemplate(file string) string {
	asset, err := assets.Asset(file)
	if err != nil {
		return ""
	}
	return string(asset)
}

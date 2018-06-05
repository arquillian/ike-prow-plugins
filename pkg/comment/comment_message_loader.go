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
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
)

// Comment struct for the plugin comment service
type Comment struct {
	Message *Message
	Log     log.Logger
}

// Message struct for message template data for the plugin comment service
type Message struct {
	Thumbnail      string
	Description    string
	ConfigFile     string
	Documentation  string
	MessageFileURL string
}

// LoadMessage loads a comment message for the plugins from the template files. If the comment message is set in config then it takes that one, the default otherwise.
func (comment *Comment) LoadMessage(configuration config.PluginConfiguration, change scm.RepositoryChange) string {
	var msg string
	if configuration.PluginHint != "" {
		msg = comment.getMsgFromConfigHint(configuration, change)
	} else if content := comment.defaultFileContent(configuration, change); content != "" {
		msg = content
	} else if configuration.LocationURL == "" {
		msg = comment.loadMessageTemplate("message-with-no-config.txt")
	} else {
		msg = comment.loadMessageTemplate("message-with-config.txt")
	}
	return comment.getMsgFromTemplate(msg)
}

func (comment *Comment) getMsgFromTemplate(msg string) string {
	var tpl bytes.Buffer
	msgTemplate, err := template.New("message").Parse(msg)
	if err != nil {
		comment.Log.Errorf("falied to parse template file %s", err)
	}
	err = msgTemplate.Execute(&tpl, comment.Message)
	if err != nil {
		comment.Log.Errorf("failed to write template output %s", err)
	}
	return tpl.String()
}

func (comment *Comment) getMsgFromConfigHint(configuration config.PluginConfiguration, change scm.RepositoryChange) string {
	fileRegex := "(?mi)" + configuration.PluginName + "_hint.md$"

	isFilePath, err := regexp.MatchString(fileRegex, configuration.PluginHint)
	if isFilePath && err == nil {
		return comment.getMsgFromFile(configuration, change)
	}
	return configuration.PluginHint
}

func (comment *Comment) getMsgFromFile(configuration config.PluginConfiguration, change scm.RepositoryChange) string {
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
		comment.Message.MessageFileURL = msgFileURL
		return comment.loadMessageTemplate("message-with-hint-file-not-found.txt")
	}

	return string(content)
}

func (comment *Comment) defaultFileContent(configuration config.PluginConfiguration, change scm.RepositoryChange) string {
	pluginHintPath := fmt.Sprintf("%s%s_hint.md", ghservice.ConfigHome, configuration.PluginName)
	ghFileService := ghservice.RawFileService{Change: change}
	hintURL := ghFileService.GetRawFileURL(pluginHintPath)

	content, e := utils.GetFileFromURL(hintURL)
	if e != nil {
		return ""
	}
	return string(content)
}

func (comment *Comment) loadMessageTemplate(file string) string {
	asset, err := assets.Asset(file)
	if err != nil {
		comment.Log.Errorf("failed to load template asset file %s", err)
	}
	return string(asset)
}

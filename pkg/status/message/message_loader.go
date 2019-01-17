package message

import (
	"bytes"
	"fmt"
	"text/template"

	assets "github.com/arquillian/ike-prow-plugins/pkg/assets/generated"
	ghservice "github.com/arquillian/ike-prow-plugins/pkg/github/service"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
)

// Loader keeps information necessary for status message loading
type Loader struct {
	Message    *Message
	Log        log.Logger
	PluginName string
}

// Message keeps all data used in message templates
type Message struct {
	Thumbnail      string
	Description    string
	ConfigFile     string
	Documentation  string
	MessageFileURL string
}

// LoadMessage loads a status message from the template files
func (l *Loader) LoadMessage(change scm.RepositoryChange, statusFileSpec string) string {
	var msg string

	if content := l.defaultFileContent(l.PluginName, change, statusFileSpec); content != "" {
		msg = content
	} else if l.Message.ConfigFile == "" {
		msg = l.loadMessageTemplate("message-with-no-config.txt")
	} else {
		msg = l.loadMessageTemplate("message-with-config.txt")
	}
	return l.getMsgFromTemplate(msg)
}

func (l *Loader) getMsgFromTemplate(msg string) string {
	var tpl bytes.Buffer
	msgTemplate, err := template.New("message").Parse(msg)
	if err != nil {
		l.Log.Errorf("failed to parse template file %s", err)
	}
	err = msgTemplate.Execute(&tpl, l.Message)
	if err != nil {
		l.Log.Errorf("failed to write template output %s", err)
	}
	return tpl.String()
}

func (l *Loader) defaultFileContent(pluginName string, change scm.RepositoryChange, defaultFileSpec string) string {
	if defaultFileSpec != "" {
		defaultFileSpec = "_" + defaultFileSpec
	}
	statusMsgPath := fmt.Sprintf("%s%s%s_message.md", ghservice.ConfigHome, pluginName, defaultFileSpec)
	ghFileService := ghservice.RawFileService{Change: change}

	content, e := utils.GetFileFromURL(ghFileService.GetRawFileURL(statusMsgPath))
	if e != nil {
		return ""
	}
	return string(content)
}

func (l *Loader) loadMessageTemplate(file string) string {
	asset, err := assets.Asset(file)
	if err != nil {
		l.Log.Errorf("failed to load template asset file %s", err)
	}
	return string(asset)
}

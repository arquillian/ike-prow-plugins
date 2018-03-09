package main

import (
	"github.com/sirupsen/logrus"
	gogh "github.com/google/go-github/github"
	"k8s.io/test-infra/prow/pluginhelp"

	"github.com/arquillian/ike-prow-plugins/pkg/plugin/server"
	pluginBootstrap "github.com/arquillian/ike-prow-plugins/pkg/plugin"
	"github.com/arquillian/ike-prow-plugins/pkg/github"
)

// ProwPluginName is an external prow plugin name used to register this service
const ProwPluginName = "pr-sanitizer"

var (
	log = logrus.StandardLogger().WithField("ike-plugins", ProwPluginName)
)

// GitHubLabelsEventsHandler is the event handler for the plugin.
// Implements server.GitHubEventHandler interface which contains the logic for incoming GitHub events
type GitHubLabelsEventsHandler struct {
	Client *gogh.Client
	log    *logrus.Entry
}

func main() {
	pluginBootstrap.InitPlugin(log, handlerCreator, serverCreator, helpProvider)
}

func handlerCreator(githubClient *gogh.Client) server.GitHubEventHandler {
	return &GitHubLabelsEventsHandler{
		Client: githubClient,
	}
}

func serverCreator(webhookSecret []byte, eventHandler server.GitHubEventHandler) (*server.Server) {
	return &server.Server{
		GitHubEventHandler: eventHandler,
		HmacSecret:         webhookSecret,
		Log:                log,
	}
}

// HandleEvent is an entry point for the plugin logic. This method is invoked by the Server when
// events are dispatched from the /hook service
func (gh *GitHubLabelsEventsHandler) HandleEvent(eventType github.EventType, eventGUID string, payload []byte) error {
	gh.log = logrus.StandardLogger().WithField("ike-plugins", ProwPluginName).WithFields(
		logrus.Fields{
			"event-type": eventType,
			"event-GUID": eventGUID,
		},
	)

	gh.log.Info("Handling labels event.")
	return nil
}

func helpProvider(enabledRepos []string) (*pluginhelp.PluginHelp, error) {
	return &pluginhelp.PluginHelp{
		Description: `PR Sanitizer plugin`,
	}, nil
}

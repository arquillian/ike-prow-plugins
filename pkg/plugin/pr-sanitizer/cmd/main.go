package main

import (
	"k8s.io/test-infra/prow/pluginhelp"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	pluginBootstrap "github.com/arquillian/ike-prow-plugins/pkg/plugin"
	"github.com/arquillian/ike-prow-plugins/pkg/server"
)

// ProwPluginName is an external prow plugin name used to register this service
const ProwPluginName = "pr-sanitizer"

// GitHubLabelsEventsHandler is the event handler for the plugin.
// Implements server.GitHubEventHandler interface which contains the logic for incoming GitHub events
type GitHubLabelsEventsHandler struct {
	Client  github.Client
	BotName string
}

func main() {
	pluginBootstrap.InitPlugin(ProwPluginName, handlerCreator, serverCreator, helpProvider)
}

func handlerCreator(githubClient github.Client, botName string) server.GitHubEventHandler {
	return &GitHubLabelsEventsHandler{Client: githubClient, BotName: botName}
}

func serverCreator(webhookSecret []byte, eventHandler server.GitHubEventHandler) *server.Server {
	return &server.Server{
		GitHubEventHandler: eventHandler,
		HmacSecret:         webhookSecret,
	}
}

// HandleEvent is an entry point for the plugin logic. This method is invoked by the Server when
// events are dispatched from the /hook service
func (gh *GitHubLabelsEventsHandler) HandleEvent(log log.Logger, eventType github.EventType, payload []byte) error {

	return nil
}

func helpProvider(enabledRepos []string) (*pluginhelp.PluginHelp, error) {
	return &pluginhelp.PluginHelp{
		Description: `PR Sanitizer plugin`,
	}, nil
}

package main

import (
	"github.com/arquillian/ike-prow-plugins/pkg/server"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper/plugin"
	gogh "github.com/google/go-github/github"
	"k8s.io/test-infra/prow/pluginhelp"

	pluginBootstrap "github.com/arquillian/ike-prow-plugins/pkg/plugin"
)

func main() {
	pluginBootstrap.InitPlugin(plugin.ProwPluginName, eventHandler, eventServer, helpProvider)
}

func eventHandler(githubClient *gogh.Client) server.GitHubEventHandler {
	return &plugin.GitHubTestEventsHandler{Client: githubClient}
}

func eventServer(webhookSecret []byte, eventHandler server.GitHubEventHandler) *server.Server {
	return &server.Server{
		GitHubEventHandler: eventHandler,
		HmacSecret:         webhookSecret,
	}
}

func helpProvider(enabledRepos []string) (*pluginhelp.PluginHelp, error) {
	return &pluginhelp.PluginHelp{
		Description: `Test Keeper plugin`,
	}, nil
}

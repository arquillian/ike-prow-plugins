package main

import (
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper"
	"github.com/arquillian/ike-prow-plugins/pkg/server"
	"k8s.io/test-infra/prow/pluginhelp"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	pluginBootstrap "github.com/arquillian/ike-prow-plugins/pkg/plugin"
)

func main() {
	pluginBootstrap.InitPlugin(test_keeper.ProwPluginName, eventHandler, eventServer, helpProvider)
}

func eventHandler(githubClient *github.Client) server.GitHubEventHandler {
	return &test_keeper.GitHubTestEventsHandler{Client: githubClient}
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

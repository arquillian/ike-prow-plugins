package main

import (
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper"
	"github.com/arquillian/ike-prow-plugins/pkg/server"
	"k8s.io/test-infra/prow/pluginhelp"

	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
	pluginBootstrap "github.com/arquillian/ike-prow-plugins/pkg/plugin"
)

func main() {
	pluginBootstrap.InitPlugin(testkeeper.ProwPluginName, eventHandler, eventServer, helpProvider)
}

func eventHandler(githubClient ghclient.Client, botName string) server.GitHubEventHandler {
	return &testkeeper.GitHubTestEventsHandler{Client: githubClient, BotName: botName}
}

func eventServer(webhookSecret []byte, eventHandler server.GitHubEventHandler) (*server.Server, []error) {
	errors := testkeeper.RegisterMetrics()

	return &server.Server{
		GitHubEventHandler: eventHandler,
		HmacSecret:         webhookSecret,
	}, errors
}

func helpProvider(enabledRepos []string) (*pluginhelp.PluginHelp, error) {
	return &pluginhelp.PluginHelp{
		Description: `Test Keeper plugin`,
	}, nil
}

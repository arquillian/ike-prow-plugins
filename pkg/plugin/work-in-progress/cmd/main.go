package main

import (
	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
	pluginBootstrap "github.com/arquillian/ike-prow-plugins/pkg/plugin"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/work-in-progress"
	"github.com/arquillian/ike-prow-plugins/pkg/server"
	"k8s.io/test-infra/prow/pluginhelp"
)

func main() {
	pluginBootstrap.InitPlugin(wip.ProwPluginName, handlerCreator, serverCreator, helpProvider, registerMetrics)
}

func handlerCreator(githubClient ghclient.Client, botName string) server.GitHubEventHandler {
	return &wip.GitHubWIPPRHandler{Client: githubClient, BotName: botName}
}

func serverCreator(webhookSecret []byte, eventHandler server.GitHubEventHandler) *server.Server {
	return &server.Server{
		GitHubEventHandler: eventHandler,
		HmacSecret:         webhookSecret,
	}
}

func helpProvider(enabledRepos []string) (*pluginhelp.PluginHelp, error) {
	return &pluginhelp.PluginHelp{
		Description: `Work-in-progress plugin`,
	}, nil
}

func registerMetrics() []error {
	return make([]error, 0)
}

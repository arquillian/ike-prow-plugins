package main

import (
	"k8s.io/test-infra/prow/pluginhelp"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	pluginBootstrap "github.com/arquillian/ike-prow-plugins/pkg/plugin"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/work-in-progress"
	"github.com/arquillian/ike-prow-plugins/pkg/server"
)

func main() {
	pluginBootstrap.InitPlugin(wip.ProwPluginName, handlerCreator, serverCreator, helpProvider)
}

func handlerCreator(githubClient github.Client, botName string) server.GitHubEventHandler {
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

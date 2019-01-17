package main

import (
	ghclient "github.com/arquillian/ike-prow-plugins/pkg/github/client"
	pluginBootstrap "github.com/arquillian/ike-prow-plugins/pkg/plugin"
	wip "github.com/arquillian/ike-prow-plugins/pkg/plugin/work-in-progress"
	"github.com/arquillian/ike-prow-plugins/pkg/server"
	"k8s.io/test-infra/prow/pluginhelp"
)

func main() {
	pluginBootstrap.InitPlugin(wip.ProwPluginName, handlerCreator, serverCreator, helpProvider)
}

func handlerCreator(githubClient ghclient.Client, botName string) server.GitHubEventHandler {
	return &wip.GitHubWIPPRHandler{Client: githubClient, BotName: botName}
}

func serverCreator(webhookSecret []byte, eventHandler server.GitHubEventHandler) (*server.Server, []error) {
	return &server.Server{
		GitHubEventHandler: eventHandler,
		HmacSecret:         webhookSecret,
	}, nil
}

func helpProvider(_ []string) (*pluginhelp.PluginHelp, error) { // nolint:unparam
	return &pluginhelp.PluginHelp{
		Description: `Work-in-progress plugin`,
	}, nil
}

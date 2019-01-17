package main

import (
	"k8s.io/test-infra/prow/pluginhelp"

	ghclient "github.com/arquillian/ike-prow-plugins/pkg/github/client"
	pluginBootstrap "github.com/arquillian/ike-prow-plugins/pkg/plugin"
	prsanitizer "github.com/arquillian/ike-prow-plugins/pkg/plugin/pr-sanitizer"
	"github.com/arquillian/ike-prow-plugins/pkg/server"
)

func main() {
	pluginBootstrap.InitPlugin(prsanitizer.ProwPluginName, handlerCreator, serverCreator, helpProvider)
}

func handlerCreator(githubClient ghclient.Client, botName string) server.GitHubEventHandler {
	return &prsanitizer.GitHubPRSanitizerEventsHandler{Client: githubClient, BotName: botName}
}

func serverCreator(webhookSecret []byte, eventHandler server.GitHubEventHandler) (*server.Server, []error) {
	return &server.Server{
		GitHubEventHandler: eventHandler,
		HmacSecret:         webhookSecret,
	}, nil
}

func helpProvider(_ []string) (*pluginhelp.PluginHelp, error) { // nolint:unparam
	return &pluginhelp.PluginHelp{
		Description: `PR Sanitizer plugin`,
	}, nil
}

package main

import (
	"github.com/google/go-github/github"
	"k8s.io/test-infra/prow/pluginhelp"

	pluginBootstrap "github.com/arquillian/ike-prow-plugins/pkg/plugin"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/work-in-progress/plugin"
	"github.com/arquillian/ike-prow-plugins/pkg/server"
)

func main() {
	pluginBootstrap.InitPlugin(plugin.ProwPluginName, handlerCreator, serverCreator, helpProvider)
}

func handlerCreator(githubClient *github.Client) server.GitHubEventHandler {
	return &plugin.GitHubWIPPRHandler{Client: githubClient}
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

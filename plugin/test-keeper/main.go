package main

import (
	"github.com/sirupsen/logrus"
	"github.com/google/go-github/github"
	"k8s.io/test-infra/prow/pluginhelp"

	"github.com/arquillian/ike-prow-plugins/plugin/server"
	"github.com/arquillian/ike-prow-plugins/plugin/test-keeper/plugin"

	pluginBootstrap "github.com/arquillian/ike-prow-plugins/plugin"
)

var (
	log = logrus.StandardLogger().WithField("ike-plugins", plugin.ProwPluginName)
)

func main() {
	pluginBootstrap.InitPlugin(log, handlerCreator, serverCreator, helpProvider)
}

func handlerCreator(githubClient *github.Client) server.GitHubEventHandler {
	return &plugin.GitHubTestEventsHandler{
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

func helpProvider(enabledRepos []string) (*pluginhelp.PluginHelp, error) {
	return &pluginhelp.PluginHelp{
		Description: `Test Keeper plugin`,
	}, nil
}

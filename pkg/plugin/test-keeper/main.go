package main

import (
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/server"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper/plugin"
	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/pluginhelp"

	pluginBootstrap "github.com/arquillian/ike-prow-plugins/pkg/plugin"
)

var (
	log = logrus.StandardLogger().WithField("ike-plugins", plugin.ProwPluginName)
)

func main() {
	pluginBootstrap.InitPlugin(log, eventHandler, eventServer, helpProvider)
}

func eventHandler(githubClient *github.Client) server.GitHubEventHandler {
	return &plugin.GitHubTestEventsHandler{
		Client: githubClient,
	}
}

func eventServer(webhookSecret []byte, eventHandler server.GitHubEventHandler) *server.Server {
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

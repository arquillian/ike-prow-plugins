package main


import (
	"github.com/sirupsen/logrus"
	"github.com/google/go-github/github"
	"k8s.io/test-infra/prow/pluginhelp"

	pluginBootstrap "github.com/arquillian/ike-prow-plugins/plugin"
	"github.com/arquillian/ike-prow-plugins/plugin/server"
	"github.com/arquillian/ike-prow-plugins/plugin/work-in-progress/plugin"
)

var (
	log = logrus.StandardLogger().WithField("ike-plugins", plugin.ProwPluginName)
)

func main() {
	pluginBootstrap.InitPlugin(log, handlerCreator, serverCreator, helpProvider)
}

func handlerCreator(githubClient *github.Client) server.GitHubEventHandler {
	return &plugin.GitHubWIPPRHandler{
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
		Description: `Work-in-progress plugin`,
	}, nil
}

package plugin

import (
	"flag"
	"syscall"
	"os/signal"
	"net/url"
	"golang.org/x/oauth2"
	"context"

	"github.com/sirupsen/logrus"
	"github.com/google/go-github/github"
	"k8s.io/test-infra/prow/plugins"
	"github.com/arquillian/ike-prow-plugins/plugin/server"
	"github.com/arquillian/ike-prow-plugins/plugin/utils"
	"k8s.io/test-infra/prow/pluginhelp/externalplugins"
	"strconv"

	"net/http"
)

// nolint
var (
	port              = flag.Int("port", 8888, "Port to listen on.")
	dryRun            = flag.Bool("dry-run", true, "Dry run for testing. Uses API tokens but does not mutate.")
	pluginConfig      = flag.String("ike-plugins-config", "/etc/plugins/plugins", "Path to ike-plugins config file.")
	githubEndpoint    = flag.String("github-endpoint", "https://api.github.com", "GitHub's API endpoint.")
	githubTokenFile   = flag.String("github-token-file", "/etc/github/oauth", "Path to the file containing the GitHub OAuth secret.")
	webhookSecretFile = flag.String("hmac-secret-file", "/etc/webhook/hmac", "Path to the file containing the GitHub HMAC secret.")
)

// EventHandlerCreator is a func type that creates server.GitHubEventHandler instance which is the central point for
// the plugin logic
type EventHandlerCreator func(client *github.Client) server.GitHubEventHandler

// ServerCreator is a func type that wires Server and server.GitHubEventHandler together
type ServerCreator func(hmacSecret []byte, evenHandler server.GitHubEventHandler) *server.Server

// InitPlugin instantiates logger, loads the secrets from the flags, sets context to background and starts server with
// the attached event handler.
func InitPlugin(log *logrus.Entry, eventHandlerCreator EventHandlerCreator, serverCreator ServerCreator,
	helpProvider externalplugins.ExternalPluginHelpProvider) {

	flag.Parse()
	logrus.SetFormatter(&logrus.JSONFormatter{})
	// TODO: Use global option from the prow config.
	logrus.SetLevel(logrus.InfoLevel)

	// Ignore SIGTERM so that we don't drop hooks when the pod is removed.
	// We'll get SIGTERM first and then SIGKILL after our graceful termination deadline.
	signal.Ignore(syscall.SIGTERM)

	webhookSecret := utils.LoadSecret(*webhookSecretFile)
	oauthSecret := string(utils.LoadSecret(*githubTokenFile))

	_, err := url.Parse(*githubEndpoint)
	if err != nil {
		log.WithError(err).Fatal("Must specify a valid --github-endpoint URL.")
	}

	pa := &plugins.PluginAgent{}
	if err := pa.Start(*pluginConfig); err != nil {
		log.WithError(err).Fatalf("Error loading ike-plugins config from %q.", *pluginConfig)
	}

	ctx := context.Background()
	token := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: oauthSecret},
	)
	githubClient := github.NewClient(oauth2.NewClient(ctx, token))

	handler := eventHandlerCreator(githubClient)
	pluginServer := serverCreator(webhookSecret, handler)

	log.Infof("Starting server on port %s", strconv.Itoa(*port))

	http.Handle("/", pluginServer)
	externalplugins.ServeExternalPluginHelp(http.DefaultServeMux, log, helpProvider)
	logrus.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}

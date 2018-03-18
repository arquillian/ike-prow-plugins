package plugin

import (
	"context"
	"flag"
	"net/url"
	"os/signal"
	"syscall"

	"golang.org/x/oauth2"

	"strconv"

	"github.com/arquillian/ike-prow-plugins/pkg/plugin/server"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/pluginhelp/externalplugins"
	"k8s.io/test-infra/prow/plugins"

	"net/http"

	"time"

	"github.com/evalphobia/logrus_sentry"
)

// nolint
var (
	port                = flag.Int("port", 8888, "Port to listen on.")
	dryRun              = flag.Bool("dry-run", true, "Dry run for testing. Uses API tokens but does not mutate.")
	pluginConfig        = flag.String("ike-plugins-config", "/etc/plugins/plugins", "Path to ike-plugins config file.")
	githubEndpoint      = flag.String("github-endpoint", "https://api.github.com", "GitHub's API endpoint.")
	githubTokenFile     = flag.String("github-token-file", "/etc/github/oauth", "Path to the file containing the GitHub OAuth secret.")
	webhookSecretFile   = flag.String("hmac-secret-file", "/etc/webhook/hmac", "Path to the file containing the GitHub HMAC secret.")
	sentryDsnSecretFile = flag.String("sentry-dsn-file", "/etc/sentry-dsn/sentry", "Path to the file containing the Sentry DSN url.")
	sentryTimeout       = flag.Int64("sentry-timeout", 1000, "Sentry server timeout in ms. Defaults to 1 second ")
)

// EventHandlerCreator is a func type that creates server.GitHubEventHandler instance which is the central point for
// the plugin logic
type EventHandlerCreator func(client *github.Client) server.GitHubEventHandler

// ServerCreator is a func type that wires Server and server.GitHubEventHandler together
type ServerCreator func(hmacSecret []byte, evenHandler server.GitHubEventHandler) *server.Server

// InitPlugin instantiates logger, loads the secrets from the flags, sets context to background and starts server with
// the attached event handler.
func InitPlugin(pluginName string, newEventHandler EventHandlerCreator, newServer ServerCreator,
	helpProvider externalplugins.ExternalPluginHelpProvider) {

	// Ignore SIGTERM so that we don't drop hooks when the pod is removed.
	// We'll get SIGTERM first and then SIGKILL after our graceful termination deadline.
	signal.Ignore(syscall.SIGTERM)

	flag.Parse()

	log := configureLogrus(pluginName)

	webhookSecret, err := utils.LoadSecret(*webhookSecretFile)
	if err != nil {
		log.WithError(err).Fatalf("unable to load webhook secret from %q", *webhookSecretFile)
	}

	oauthSecret, err := utils.LoadSecret(*githubTokenFile)
	if err != nil {
		log.WithError(err).Fatalf("unable to load oauth token from %q", *githubTokenFile)
	}

	_, err = url.Parse(*githubEndpoint)
	if err != nil {
		log.WithError(err).Fatalf("Must specify a valid --github-endpoint URL.")
	}

	pa := &plugins.PluginAgent{}
	if err := pa.Start(*pluginConfig); err != nil {
		log.WithError(err).Fatalf("Error loading ike-plugins config from %q.", *pluginConfig)
	}

	ctx := context.Background()
	token := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: string(oauthSecret)},
	)
	githubClient := github.NewClient(oauth2.NewClient(ctx, token))

	handler := newEventHandler(githubClient)
	pluginServer := newServer(webhookSecret, handler)

	port := strconv.Itoa(*port)
	log.Infof("Starting server on port %s", port)

	http.Handle("/", pluginServer)
	externalplugins.ServeExternalPluginHelp(http.DefaultServeMux, log, helpProvider)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.WithError(err).Fatalf("failed to start server on port %s", port)
	}
}

func configureLogrus(pluginName string) *logrus.Entry {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.WarnLevel)

	log := logrus.WithField("ike-plugins", pluginName)

	sentryDsn, err := utils.LoadSecret(*sentryDsnSecretFile)
	if err != nil {
		log.WithError(err).Errorf("unable to load sentry dsn from %q. No sentry integration enabled", sentryDsnSecretFile)
	}

	if sentryDsn != nil {
		tags := map[string]string{
			"plugin": pluginName,
		}

		levels := []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
		}

		hook, err := logrus_sentry.NewWithTagsSentryHook(string(sentryDsn), tags, levels)

		if err == nil {
			hook.Timeout = time.Duration(*sentryTimeout) * time.Second
			logrus.AddHook(hook)
		} else {
			log.WithError(err).Error("failed to add sentry hook")
		}
	}

	return log
}

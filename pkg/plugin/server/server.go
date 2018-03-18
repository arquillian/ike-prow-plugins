package server

import (
	"net/http"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/sirupsen/logrus"
	"k8s.io/test-infra/prow/hook"
)

// GitHubEventHandler is a type which keeps the logic of handling GitHub events for the given plugin implementation.
// It is used by Server implementation to handle incoming events.
type GitHubEventHandler interface {
	HandleEvent(log log.Logger, eventType github.EventType, payload []byte) error
}

// Server implements http.Handler. It validates incoming GitHub webhooks and
// then dispatches them to the appropriate plugins.
type Server struct {
	GitHubEventHandler GitHubEventHandler
	HmacSecret         []byte
	PluginName         string
}

// ServeHTTP validates an incoming webhook and puts it into the event channel.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO(k8s-prow): Move webhook handling logic out of hook binary so that we don't have to import all
	eventType, eventGUID, payload, ok := hook.ValidateWebhook(w, r, s.HmacSecret)
	if !ok {
		return
	}

	l := logrus.StandardLogger().WithFields(
		logrus.Fields{
			"ike-plugins": s.PluginName,
			"event-type":  eventType,
			"event-GUID":  eventGUID,
		},
	)

	if err := s.GitHubEventHandler.HandleEvent(l, github.EventType(eventType), payload); err != nil {
		l.WithError(err).Error("error parsing event.")
		return
	}
}

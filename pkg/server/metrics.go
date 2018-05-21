package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/google/go-github/github"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
)

var (
	namespace = ""
	subsystem = "ike-prow-plugins"

	rateLimit = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "github_rate_limits",
		Help: "Rate limit for GitHub API calls",
	}, []string{"type"})

	webhookCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "webhook_counter",
		Help: "A counter of the webhooks made to service.",
	}, []string{"full_name"})

	handledEventsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "handled_events_counter",
		Help: "A counter of handled events.",
	}, []string{"event_type"})
)

// RegisterMetrics registers prometheus collectors to collect metrics
func RegisterMetrics(log log.Logger) {
	rateLimit = register(rateLimit, "github_rate_limits", log).(*prometheus.GaugeVec)
	webhookCounter = register(webhookCounter, "webhook_counter", log).(*prometheus.CounterVec)
	handledEventsCounter = register(handledEventsCounter, "handled_events_counter", log).(*prometheus.CounterVec)
}

func register(c prometheus.Collector, name string, log log.Logger) prometheus.Collector {
	err := prometheus.Register(c)
	if err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			return are.ExistingCollector
		}

		log.Panic(map[string]interface{}{
			"metric_name": prometheus.BuildFQName(namespace, subsystem, name),
			"err":         err,
		}, "failed to register the prometheus metric")
	}
	return c
}

// Metrics is a set of metrics gathered by the Ike Prow Plugin.
// It includes rate limit, incoming webhooks. handled events.
type Metrics struct {
	RateLimit            *prometheus.GaugeVec
	WebhookCounter       *prometheus.CounterVec
	HandledEventsCounter *prometheus.CounterVec
}

// NewMetrics creates a new set of metrics for the Ike Prow Plugin.
func NewMetrics() *Metrics {
	return &Metrics{
		RateLimit:            rateLimit,
		WebhookCounter:       webhookCounter,
		HandledEventsCounter: handledEventsCounter,
	}
}

// ReportRateLimit reports rate limit for Github API calls.
func ReportRateLimit(r *github.RateLimits) {
	rateLimit.WithLabelValues("core").Set(float64(r.Core.Remaining))
	rateLimit.WithLabelValues("search").Set(float64(r.Search.Remaining))
}

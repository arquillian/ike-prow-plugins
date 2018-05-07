package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/google/go-github/github"
)

var (
	rateLimit = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "github_rate_limits",
		Help: "Rate limit for GitHub API calls",
	}, []string{"type"})

	webhookCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "webhook_counter",
		Help: "A counter of the webhooks made to service.",
	}, []string{"event_type"})

	handledEventsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "handled_events_counter",
		Help: "A counter of handled events.",
	}, []string{"event_type"})
)

func init() {
	prometheus.MustRegister(rateLimit)
	prometheus.MustRegister(webhookCounter)
	prometheus.MustRegister(handledEventsCounter)
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

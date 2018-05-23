package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
)

// RegisterMetrics registers prometheus collectors to collect metrics
func (s *Server) RegisterMetrics() []error {
	errors := make([]error, 0, 3)
	metrics := NewMetrics()
	if rateLimitCollector, e := registerOrGet(metrics.RateLimit); e == nil {
		metrics.RateLimit = rateLimitCollector.(*prometheus.GaugeVec)
	} else {
		errors = append(errors, e)
	}

	if webHookCollector, e := registerOrGet(metrics.WebHookCounter); e == nil {
		metrics.WebHookCounter = webHookCollector.(*prometheus.CounterVec)
	} else {
		errors = append(errors, e)
	}

	if HandledEventsCollector, e := registerOrGet(metrics.HandledEventsCounter); e == nil {
		metrics.HandledEventsCounter = HandledEventsCollector.(*prometheus.CounterVec)
	} else {
		errors = append(errors, e)
	}

	if len(errors) > 0 {
		return errors
	}

	s.Metrics = metrics
	return make([]error, 0);
}

func registerOrGet(c prometheus.Collector) (prometheus.Collector, error) {
	if err := prometheus.Register(c); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			return are.ExistingCollector, nil
		}
		return nil, err;
	}
	return c, nil
}

// Metrics is a set of metrics gathered by the Ike Prow Plugin.
// It includes rate limit, incoming webhooks. handled events.
type Metrics struct {
	RateLimit            *prometheus.GaugeVec
	WebHookCounter       *prometheus.CounterVec
	HandledEventsCounter *prometheus.CounterVec
}

// NewMetrics creates a new set of metrics for the Ike Prow Plugin.
func NewMetrics() *Metrics {
	return &Metrics{
		RateLimit: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "github_rate_limits",
			Help: "Rate limit for GitHub API calls",
		}, []string{"type"}),
		WebHookCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "webhook_counter",
			Help: "A counter of the webhooks made to service.",
		}, []string{"full_name"}),
		HandledEventsCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "handled_events_counter",
			Help: "A counter of handled events.",
		}, []string{"event_type"}),
	}
}

func (metrics *Metrics) reportRateLimit(client ghclient.Client) {
	if r := ghclient.GetRateLimits(client); r != nil {
		metrics.RateLimit.WithLabelValues("core").Set(float64(r.Core.Remaining))
		metrics.RateLimit.WithLabelValues("search").Set(float64(r.Search.Remaining))
	}
}

func (metrics *Metrics) reportIncomingWebHooks(l log.Logger, label string) {
	if counter, err := metrics.WebHookCounter.GetMetricWithLabelValues(label); err != nil {
		l.Errorf("Failed to get metric for Repository: %q. Cause: %q", label, err)
	} else {
		counter.Inc()
	}
}

func (metrics *Metrics) reportHandledEvents(l log.Logger, label string) {
	if counter, err := metrics.HandledEventsCounter.GetMetricWithLabelValues(label); err != nil {
		l.Errorf("Failed to get metric for Event: %q. Cause: %q", label, err)
	} else {
		counter.Inc()
	}
}

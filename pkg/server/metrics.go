package server

import (
	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	rateLimit = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "github_rate_limits",
		Help: "Rate limit for GitHub API calls",
	}, []string{"type"})
	webHookCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "webhook_counter",
		Help: "A counter of the webhooks made to service.",
	}, []string{"full_name"})
	handledEventsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "handled_events_counter",
		Help: "A counter of handled events.",
	}, []string{"event_type"})
)

// RegisterMetrics registers prometheus collectors to collect metrics
func RegisterMetrics(client ghclient.Client) (*Metrics, []error) {
	errors := make([]error, 0, 3)
	metrics := &Metrics{
		ghClient: client,
	}

	registerOrGet(rateLimit, &errors, func(collector prometheus.Collector) {
		metrics.RateLimit = collector.(*prometheus.GaugeVec)
	})

	registerOrGet(webHookCounter, &errors, func(collector prometheus.Collector) {
		metrics.WebHookCounter = collector.(*prometheus.CounterVec)
	})

	registerOrGet(handledEventsCounter, &errors, func(collector prometheus.Collector) {
		metrics.HandledEventsCounter = collector.(*prometheus.CounterVec)
	})

	return metrics, errors
}

func registerOrGet(collector prometheus.Collector, errors *[]error, assign func(regCollector prometheus.Collector)) {
	if err := prometheus.Register(collector); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			assign(are.ExistingCollector)
		}
		*errors = append(*errors, err)
	}
	assign(collector)
}

// Metrics is a set of metrics gathered by the Ike Prow Plugin.
// It includes rate limit, incoming webhooks. handled events.
type Metrics struct {
	RateLimit            *prometheus.GaugeVec
	WebHookCounter       *prometheus.CounterVec
	HandledEventsCounter *prometheus.CounterVec
	ghClient             ghclient.Client
}

func (m *Metrics) reportRateLimit(l log.Logger) {
	if limits, err := m.ghClient.GetRateLimit(); err != nil {
		l.Errorf("Failed to get metric GH Client rate limit. Cause: %q", err)
	} else {
		m.RateLimit.WithLabelValues("core").Set(float64(limits.Core.Remaining))
		m.RateLimit.WithLabelValues("search").Set(float64(limits.Search.Remaining))
	}
}

func (m *Metrics) reportIncomingWebHooks(l log.Logger, label string) {
	if counter, err := m.WebHookCounter.GetMetricWithLabelValues(label); err != nil {
		l.Errorf("Failed to get metric for Repository: %q. Cause: %q", label, err)
	} else {
		counter.Inc()
	}
}

func (m *Metrics) reportHandledEvents(l log.Logger, label string) {
	if counter, err := m.HandledEventsCounter.GetMetricWithLabelValues(label); err != nil {
		l.Errorf("Failed to get metric for Event: %q. Cause: %q", label, err)
	} else {
		counter.Inc()
	}
}

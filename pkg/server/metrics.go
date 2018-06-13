package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
)

var (
	rateLimit = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "github_rate_limits",
		Help: "Rate limit for GitHub API calls.",
	}, []string{"type"})
	webHookCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "webhook_total",
		Help: "Total number of the webhooks made to service.",
	}, []string{"full_name"})
	handledEventsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "handled_events_total",
		Help: "Total number of handled events.",
	}, []string{"event_type"})
	ghClient ghclient.Client
)

// RegisterMetrics registers prometheus collectors to collect metrics
func RegisterMetrics(client ghclient.Client) ([]error) {
	errors := make([]error, 0, 3)
	ghClient = client
	RegisterOrGet(rateLimit, &errors, func(collector prometheus.Collector) {
		rateLimit = collector.(*prometheus.GaugeVec)
	})

	RegisterOrGet(webHookCounter, &errors, func(collector prometheus.Collector) {
		webHookCounter = collector.(*prometheus.CounterVec)
	})

	RegisterOrGet(handledEventsCounter, &errors, func(collector prometheus.Collector) {
		handledEventsCounter = collector.(*prometheus.CounterVec)
	})

	return errors
}

// RegisterOrGet registers the provided Collector with the DefaultRegisterer and
// assigns the Collector, unless an equal Collector was registered before, in
// which case that Collector is assigned.
func RegisterOrGet(collector prometheus.Collector, errors *[]error, assign func(regCollector prometheus.Collector)) {
	if err := prometheus.Register(collector); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			assign(are.ExistingCollector)
		}
		*errors = append(*errors, err)
	}
	assign(collector)
}

func reportRateLimit(l log.Logger) {
	if limits, err := ghClient.GetRateLimit(); err != nil {
		l.Errorf("Failed to get metric GH Client rate limit. Cause: %q", err)
	} else {
		rateLimit.WithLabelValues("core").Set(float64(limits.Core.Remaining))
		rateLimit.WithLabelValues("search").Set(float64(limits.Search.Remaining))
	}
}

func reportIncomingWebHooks(l log.Logger, label string) {
	if counter, err := webHookCounter.GetMetricWithLabelValues(label); err != nil {
		l.Errorf("Failed to get metric for Repository: %q. Cause: %q", label, err)
	} else {
		counter.Inc()
	}
}

func reportHandledEvents(l log.Logger, label string) {
	if counter, err := handledEventsCounter.GetMetricWithLabelValues(label); err != nil {
		l.Errorf("Failed to get metric for Event: %q. Cause: %q", label, err)
	} else {
		counter.Inc()
	}
}

// RateLimitWithLabelValues replaces the method of the same name in MetricVec.
func RateLimitWithLabelValues(lvs ...string) (prometheus.Gauge, error) {
	return rateLimit.GetMetricWithLabelValues(lvs...)
}

// WebHookCounterWithLabelValues replaces the method of the same name in MetricVec.
func WebHookCounterWithLabelValues(lvs ...string) (prometheus.Counter, error) {
	return webHookCounter.GetMetricWithLabelValues(lvs...)
}

// HandledEventsCounterWithLabelValues replaces the method of the same name in MetricVec.
func HandledEventsCounterWithLabelValues(lvs ...string) (prometheus.Counter, error) {
	return handledEventsCounter.GetMetricWithLabelValues(lvs...)
}

// UnRegisterAndResetMetrics unregisters and reset prometheus collectors.
func UnRegisterAndResetMetrics() {
	prometheus.Unregister(webHookCounter)
	webHookCounter.Reset()
	prometheus.Unregister(rateLimit)
	rateLimit.Reset()
	prometheus.Unregister(handledEventsCounter)
	handledEventsCounter.Reset()
}

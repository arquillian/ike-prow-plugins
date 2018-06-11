package testkeeper

import (
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	gogh "github.com/google/go-github/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/arquillian/ike-prow-plugins/pkg/server"
)

const (
	// WithTests is a label used in prometheus metrics pull_requests_total for pull request containing some tests.
	WithTests = "with_tests"
	// WithoutTests is a label used in prometheus metrics pull_requests_total for pull request without any tests.
	WithoutTests = "without_tests"
)

var (
	pullRequestsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "pull_requests_total",
		Help: "Total number of pull requests.",
	}, []string{"full_name", "type"})
	okWithoutTestsPullRequest = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "pull_requests_changed_file_size",
		Help:    "Histogram for changed file size of pull request with ok-without-tests.",
		Buckets: prometheus.ExponentialBuckets(1, 5, 6),
	}, []string{"full_name"})
)

// RegisterMetrics registers prometheus collectors to collect metrics for test-keeper.
func RegisterMetrics() []error {
	errors := make([]error, 0, 2)
	server.RegisterOrGet(pullRequestsCounter, &errors, func(collector prometheus.Collector) {
		pullRequestsCounter = collector.(*prometheus.CounterVec)
	})

	server.RegisterOrGet(okWithoutTestsPullRequest, &errors, func(collector prometheus.Collector) {
		okWithoutTestsPullRequest = collector.(*prometheus.HistogramVec)
	})

	return errors
}

func reportPullRequest(l log.Logger, pr *gogh.PullRequest, prType string) {
	fullName := *pr.Head.Repo.FullName
	if counter, err := pullRequestsCounter.GetMetricWithLabelValues(fullName, prType); err != nil {
		l.Errorf("Failed to get pull request metric for Repository: %q. Cause: %q", fullName, err)
	} else {
		counter.Inc()
	}
}

func reportOkWithoutTestsPullRequest(pr *gogh.PullRequest) {
	okWithoutTestsPullRequest.WithLabelValues(*pr.Head.Repo.FullName).Observe(float64(*pr.ChangedFiles))
}

// PullRequestCounterWithLabelValues replaces the method of the same name in MetricVec.
func PullRequestCounterWithLabelValues(lvs ...string) (prometheus.Counter, error) {
	return pullRequestsCounter.GetMetricWithLabelValues(lvs...)
}

// OkWithoutTestsPullRequestWithLabelValues replaces the method of the same name in MetricVec.
func OkWithoutTestsPullRequestWithLabelValues(lvs ...string) (prometheus.Observer, error) {
	return okWithoutTestsPullRequest.GetMetricWithLabelValues(lvs...)
}

// UnRegisterMetricsAndReset unregisters and reset prometheus collectors.
func UnRegisterMetricsAndReset() {
	pullRequestsCounter.Reset()
	prometheus.Unregister(pullRequestsCounter)
	okWithoutTestsPullRequest.Reset()
	prometheus.Unregister(okWithoutTestsPullRequest)
}

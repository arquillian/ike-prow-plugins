package testkeeper

import (
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/server"
	gogh "github.com/google/go-github/v41/github"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// WithTests is a label used in prometheus metrics pull_requests_total for pull request containing some tests.
	WithTests = "with_tests"
	// WithoutTests is a label used in prometheus metrics pull_requests_total for pull request without any tests.
	WithoutTests = "without_tests"
)

var (
	pullRequestsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "test_keeper_pull_request_total",
		Help: "Total number of pull requests received in test-keeper plugin",
	}, []string{"full_name", "type"})
	okWithoutTestsPullRequest = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "test_keeper_pull_requests_size",
		Help:    "Histogram for pull request size with ok-without-tests command applied in test-keeper plugin",
		Buckets: prometheus.ExponentialBuckets(1, 3, 6),
	}, []string{"full_name"})
)

// RegisterMetrics registers prometheus collectors to collect metrics for test-keeper.
func RegisterMetrics() []error {
	errors := make([]error, 0, 2)
	server.RegisterOrAssignCollector(pullRequestsCounter, &errors, func(collector prometheus.Collector) {
		pullRequestsCounter = collector.(*prometheus.CounterVec)
	})

	server.RegisterOrAssignCollector(okWithoutTestsPullRequest, &errors, func(collector prometheus.Collector) {
		okWithoutTestsPullRequest = collector.(*prometheus.HistogramVec)
	})

	return errors
}

func reportPullRequest(l log.Logger, pr *gogh.PullRequest, prType string) {
	fullName := *pr.Base.Repo.FullName
	if counter, err := pullRequestsCounter.GetMetricWithLabelValues(fullName, prType); err != nil {
		l.Errorf("Failed to get pull request metric for Repository: %q. Cause: %q", fullName, err)
	} else {
		counter.Inc()
	}
}

func reportBypassCommand(pr *gogh.PullRequest) {
	okWithoutTestsPullRequest.WithLabelValues(*pr.Base.Repo.FullName).Observe(float64(*pr.ChangedFiles))
}

// PullRequestCounterWithLabelValues replaces the method of the same name in MetricVec.
func PullRequestCounterWithLabelValues(lvs ...string) (prometheus.Counter, error) {
	return pullRequestsCounter.GetMetricWithLabelValues(lvs...)
}

// OkWithoutTestsPullRequestWithLabelValues replaces the method of the same name in MetricVec.
func OkWithoutTestsPullRequestWithLabelValues(lvs ...string) (prometheus.Observer, error) {
	return okWithoutTestsPullRequest.GetMetricWithLabelValues(lvs...)
}

// UnRegisterAndResetMetrics unregisters and reset prometheus collectors.
func UnRegisterAndResetMetrics() {
	pullRequestsCounter.Reset()
	prometheus.Unregister(pullRequestsCounter)
	okWithoutTestsPullRequest.Reset()
	prometheus.Unregister(okWithoutTestsPullRequest)
}

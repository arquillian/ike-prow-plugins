package server_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http/httptest"
	"github.com/arquillian/ike-prow-plugins/pkg/server"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"gopkg.in/h2non/gock.v1"
	"k8s.io/test-infra/prow/phony"
	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
)

type DummyGHEventHandler struct {
}

func (gh *DummyGHEventHandler) HandleEvent(log log.Logger, eventType github.EventType, payload []byte) error {
	return nil
}

var _ = Describe("Service Metrics", func() {
	secret := []byte("123abc")
	client := test.NewDefaultGitHubClient()
	var (
		serverMetrics *server.Metrics
		testServer    *httptest.Server
	)

	BeforeEach(func() {
		prowServer := &server.Server{
			GitHubEventHandler: &DummyGHEventHandler{},
			PluginName:         "dummy-name",
			HmacSecret:         secret,
		}
		metrics, errs := server.RegisterMetrics(client)
		if len(errs) > 0 {
			var msg string
			for _, er := range errs {
				msg += er.Error() + "\n"
			}
			Fail("Prometheus serverMetrics registration failed with errors:\n" + msg)
		}
		prowServer.Metrics = metrics
		serverMetrics = metrics
		testServer = httptest.NewServer(prowServer)
	})

	AfterEach(func() {
		testServer.Close()
		prometheus.Unregister(serverMetrics.WebHookCounter)
		prometheus.Unregister(serverMetrics.RateLimit)
		prometheus.Unregister(serverMetrics.HandledEventsCounter)
	})

	It("should count incoming webhook", func() {
		// given
		fullName := "bartoszmajsak/wfswarm-booster-pipeline-test"
		payload := test.LoadFromFile("../plugin/work-in-progress/test_fixtures/github_calls/ready_pr_opened.json")

		// when
		err := phony.SendHook(testServer.URL, "pull_request", payload, secret)

		// then
		Ω(err).ShouldNot(HaveOccurred())
		counter, err := serverMetrics.WebHookCounter.GetMetricWithLabelValues(fullName)
		Ω(err).ShouldNot(HaveOccurred())
		Expect(count(counter)).To(Equal(1))
	})

	It("should count handled events", func() {
		// given
		eventType := "issue_comment"
		payload := test.LoadFromFile("../plugin/test-keeper/test_fixtures/github_calls/prs/without_tests/skip_comment_by_admin.json")

		// when
		err := phony.SendHook(testServer.URL, "issue_comment", payload, secret)

		// then
		Ω(err).ShouldNot(HaveOccurred())
		counter, err := serverMetrics.HandledEventsCounter.GetMetricWithLabelValues(eventType)
		Ω(err).ShouldNot(HaveOccurred())
		Expect(count(counter)).To(Equal(1))
	})

	It("should get Rate limit for GitHub API calls", func() {
		// given
		defer gock.DisableNetworking()
		gock.New(testServer.URL).
			EnableNetworking()

		gock.New("https://api.github.com").
			Get("/rate_limit").
			Persist().
			Reply(200).
			Body(test.FromFile("../github/client/test_fixtures/gh/low_rate_limit.json"))

		payload := test.LoadFromFile("../plugin/test-keeper/test_fixtures/github_calls/prs/without_tests/skip_comment_by_admin.json")

		// when
		err := phony.SendHook(testServer.URL, "issue_comment", payload, secret)

		// then
		Ω(err).ShouldNot(HaveOccurred())
		
		gauge, err := serverMetrics.RateLimit.GetMetricWithLabelValues("core")
		Ω(err).ShouldNot(HaveOccurred())
		Expect(gaugeValue(gauge)).To(Equal(8))

		gauge, err = serverMetrics.RateLimit.GetMetricWithLabelValues("search")
		Ω(err).ShouldNot(HaveOccurred())
		Expect(gaugeValue(gauge)).To(Equal(10))
	})
})

func count(counter prometheus.Counter) int {
	m := &dto.Metric{}
	counter.Write(m)
	return int(m.Counter.GetValue())
}

func gaugeValue(gauge prometheus.Gauge) int {
	m := &dto.Metric{}
	gauge.Write(m)
	return int(m.Gauge.GetValue())
}

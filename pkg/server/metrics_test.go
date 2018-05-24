package server_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http/httptest"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper"
	"github.com/arquillian/ike-prow-plugins/pkg/server"
	dto "github.com/prometheus/client_model/go"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"gopkg.in/h2non/gock.v1"
	"k8s.io/test-infra/prow/phony"
)

var _ = Describe("Service Metrics", func() {

	logger := log.NewTestLogger()
	secret := []byte("123abc")
	client := test.NewDefaultGitHubClient()
	pluginName := testkeeper.ProwPluginName
	var (
		metrics *server.Metrics
		s       *httptest.Server
	)

	BeforeEach(func() {
		eventHandler := &testkeeper.GitHubTestEventsHandler{Client: client, BotName: "alien-ike"}
		prowServer := &server.Server{
			GitHubEventHandler: eventHandler,
			PluginName:         pluginName,
			HmacSecret:         secret,
		}
		m, errs := server.RegisterMetrics(client);
		if  len(errs) > 0 {
			logger.Errorf("Prometheus metrics registration failed!")
		} else {
			prowServer.Metrics = m
			metrics = m
		}
		s = httptest.NewServer(prowServer)
	})

	AfterEach(func() {
		s.Close()
	})

	It("should count incoming webhook", func() {
		// given
		fullName := "bartoszmajsak/wfswarm-booster-pipeline-test"
		payload := test.LoadFromFile("../plugin/work-in-progress/test_fixtures/github_calls/ready_pr_opened.json")

		if err := phony.SendHook(s.URL, "pull_request", payload, secret); err != nil {
			logger.Errorf("Error sending hook: %v", err)
		}

		// when
		counter, err := metrics.WebHookCounter.GetMetricWithLabelValues(fullName)

		// then
		Ω(err).ShouldNot(HaveOccurred())
		Expect(count(counter)).To(Equal(1))
	})

	It("should count handled events", func() {
		// given
		eventType := "issue_comment"
		payload := test.LoadFromFile("../plugin/test-keeper/test_fixtures/github_calls/prs/without_tests/skip_comment_by_admin.json")

		if err := phony.SendHook(s.URL, "issue_comment", payload, secret); err != nil {
			logger.Errorf("Error sending hook: %v", err)
		}

		// when
		counter, err := metrics.HandledEventsCounter.GetMetricWithLabelValues(eventType)

		// then
		Ω(err).ShouldNot(HaveOccurred())
		Expect(count(counter)).To(Equal(1))
	})

	It("should get Rate limit for GitHub API calls", func() {
		// given
		defer gock.DisableNetworking()
		gock.New(s.URL).
			EnableNetworking()

		gock.New("https://api.github.com").
			Get("/rate_limit").
			Persist().
			Reply(200).
			Body(test.FromFile("../github/client/test_fixtures/gh/low_rate_limit.json"))

		payload := test.LoadFromFile("../plugin/test-keeper/test_fixtures/github_calls/prs/without_tests/skip_comment_by_admin.json")

		// when
		if err := phony.SendHook(s.URL, "issue_comment", payload, secret); err != nil {
			logger.Errorf("Error sending hook: %v", err)
		}

		// then
		Expect(gaugeValue(metrics.RateLimit.WithLabelValues("core"))).To(Equal(8))
		Expect(gaugeValue(metrics.RateLimit.WithLabelValues("search"))).To(Equal(10))
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

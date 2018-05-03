package server_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http/httptest"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper"
	"github.com/arquillian/ike-prow-plugins/pkg/server"
	"k8s.io/test-infra/prow/phony"
	dto "github.com/prometheus/client_model/go"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"gopkg.in/h2non/gock.v1"
	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
)

var _ = Describe("Service Metrics", func() {

	logger := log.NewTestLogger()
	secret := []byte("123abc")
	pluginName := testkeeper.ProwPluginName
	var (
		metrics *server.Metrics
		s       *httptest.Server
	)

	BeforeEach(func() {
		metrics = server.NewMetrics()
		eventHandler := &testkeeper.GitHubTestEventsHandler{Client: test.NewDefaultGitHubClient(), BotName: "alien-ike"}
		s = httptest.NewServer(&server.Server{
			GitHubEventHandler: eventHandler,
			PluginName:         pluginName,
			HmacSecret:         secret,
			Metrics:            metrics,
		})

		payload := test.LoadFromFile("../plugin/work-in-progress/test_fixtures/github_calls/ready_pr_opened.json")

		if err := phony.SendHook(s.URL, "pull_request", payload, secret); err != nil {
			logger.Errorf("Error sending hook: %v", err)
		}

		defer s.Close()
	})

	It("should count incoming webhook", func() {
		// given
		eventType := "pull_request"
		expected := 1

		// when
		counter, err := metrics.WebhookCounter.GetMetricWithLabelValues(eventType)

		// then
		Ω(err).ShouldNot(HaveOccurred())
		Expect(count(counter)).To(Equal(expected))
	})

	It("should count handled events", func() {
		// given
		eventType := "pull_request"
		expected := 2

		// when
		counter, err := metrics.HandledEventsCounter.GetMetricWithLabelValues(eventType)

		// then
		Ω(err).ShouldNot(HaveOccurred())
		Expect(count(counter)).To(Equal(expected))
	})

	It("should get Rate limit for GitHub API calls", func() {
		//given
		gock.New("https://api.github.com").
			Get("/rate_limit").
			Persist().
			Reply(200).
			Body(test.FromFile("../github/client/test_fixtures/gh/low_rate_limit.json"))

		gock.New("https://api.github.com").
			Get("/repos/owner/repo/pulls/123/files").
			Reply(200).
			BodyString("[]")

		client := ghclient.NewRateLimitWatcherClient(test.NewDefaultGitHubClient(), logger, 10)
		client.ListPullRequestFiles("owner", "repo", 123)

		// when
		core_counter := metrics.RateLimit.WithLabelValues("core")
		search_counter := metrics.RateLimit.WithLabelValues("search")


		// then
		Expect(gaugeValue(core_counter)).To(Equal(8))
		Expect(gaugeValue(search_counter)).To(Equal(10))
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
package server_test

import (
	"net/http/httptest"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper"
	"github.com/arquillian/ike-prow-plugins/pkg/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"gopkg.in/h2non/gock.v1"
	"k8s.io/test-infra/prow/phony"
)

type DummyGHEventHandler struct {
}

func (gh *DummyGHEventHandler) HandleEvent(log log.Logger, eventType github.EventType, payload []byte) error {
	return nil
}

var _ = Describe("Service Metrics", func() {
	secret := []byte("123abc")
	client := NewDefaultGitHubClient()
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
		defer gock.OffAll()
	})

	AfterEach(func() {
		testServer.Close()
		prometheus.Unregister(serverMetrics.WebHookCounter)
		serverMetrics.WebHookCounter.Reset()
		prometheus.Unregister(serverMetrics.RateLimit)
		serverMetrics.RateLimit.Reset()
		prometheus.Unregister(serverMetrics.HandledEventsCounter)
		serverMetrics.HandledEventsCounter.Reset()
		EnsureGockRequestsHaveBeenMatched()
	})

	It("should count incoming webhook", func() {
		// given
		setRateLimitMocks()
		fullName := "bartoszmajsak/wfswarm-booster-pipeline-test"
		eventPayload := MockPr(LoadedFrom("../plugin/work-in-progress/test_fixtures/github_calls/prs/pr_details.json")).
			Create().
			CreatePullRequestEvent("created")

		// when
		err := phony.SendHook(testServer.URL, string(github.PullRequest), eventPayload, secret)

		// then
		Ω(err).ShouldNot(HaveOccurred())
		counter, err := serverMetrics.WebHookCounter.GetMetricWithLabelValues(fullName)
		Ω(err).ShouldNot(HaveOccurred())
		Expect(count(counter)).To(Equal(1))
	})

	It("should count handled events", func() {
		// given
		setRateLimitMocks()
		eventPayload := MockPr(LoadedFrom("../plugin/work-in-progress/test_fixtures/github_calls/prs/pr_details.json")).
			Create().
			CreateCommentEvent(SentByRepoOwner, testkeeper.BypassCheckComment, "created")

		// when
		err := phony.SendHook(testServer.URL, string(github.IssueComment), eventPayload, secret)

		// then
		Ω(err).ShouldNot(HaveOccurred())
		counter, err := serverMetrics.HandledEventsCounter.GetMetricWithLabelValues(string(github.IssueComment))
		Ω(err).ShouldNot(HaveOccurred())
		Expect(count(counter)).To(Equal(1))
	})

	It("should get Rate limit for GitHub API calls", func() {
		// given
		setRateLimitMocks()
		eventPayload := MockPr(LoadedFrom("../plugin/work-in-progress/test_fixtures/github_calls/prs/pr_details.json")).
			Create().
			CreateCommentEvent(SentByRepoOwner, testkeeper.BypassCheckComment, "created")

		// when
		err := phony.SendHook(testServer.URL, string(github.IssueComment), eventPayload, secret)

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

func setRateLimitMocks() {
	gock.New("https://api.github.com").
		Get("/rate_limit").
		Reply(200).
		Body(FromFile("../github/client/test_fixtures/gh/low_rate_limit.json"))
	gock.New("http://127.0.0.1").
		Post("").
		EnableNetworking()
}

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

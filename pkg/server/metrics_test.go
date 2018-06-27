package server_test

import (
	"net/http/httptest"

	"encoding/json"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper"
	"github.com/arquillian/ike-prow-plugins/pkg/server"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	gogh "github.com/google/go-github/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/h2non/gock.v1"
	"k8s.io/test-infra/prow/phony"
)

type DummyGHEventHandler struct {
}

func (gh *DummyGHEventHandler) HandlePullRequestEvent(log log.Logger, event *gogh.PullRequestEvent) error {
	return nil
}

func (gh *DummyGHEventHandler) HandleIssueCommentEvent(log log.Logger, event *gogh.IssueCommentEvent) error {
	return nil
}

var _ = Describe("Service Metrics", func() {
	secret := []byte("123abc")
	client := NewDefaultGitHubClient()
	var (
		testServer *httptest.Server
	)

	BeforeEach(func() {
		prowServer := &server.Server{
			GitHubEventHandler: &DummyGHEventHandler{},
			PluginName:         "dummy-name",
			HmacSecret:         secret,
		}
		errs := server.RegisterMetrics(client)
		if len(errs) > 0 {
			var msg string
			for _, er := range errs {
				msg += er.Error() + "\n"
			}
			Fail("Prometheus serverMetrics registration failed with errors:\n" + msg)
		}
		testServer = httptest.NewServer(prowServer)
		defer gock.OffAll()
	})

	AfterEach(func() {
		testServer.Close()
		server.UnRegisterAndResetMetrics()
		EnsureGockRequestsHaveBeenMatched()
	})

	It("should count incoming webhook", func() {
		// given
		setRateLimitMocks()
		fullName := "bartoszmajsak/wfswarm-booster-pipeline-test"
		event := MockPr().
			LoadedFrom("../plugin/work-in-progress/test_fixtures/github_calls/prs/pr_details.json").
			Create().
			CreatePullRequestEvent("created")

		// when
		err := phony.SendHook(testServer.URL, string(github.PullRequest), mashal(event), secret)

		// then
		Ω(err).ShouldNot(HaveOccurred())
		counter, err := server.WebHookCounterWithLabelValues(fullName)
		Ω(err).ShouldNot(HaveOccurred())

		verifyCount(counter, 1)
	})

	It("should count handled events", func() {
		// given
		setRateLimitMocks()
		event := MockPr().
			LoadedFrom("../plugin/work-in-progress/test_fixtures/github_calls/prs/pr_details.json").
			Create().
			CreateCommentEvent(SentByRepoOwner, testkeeper.BypassCheckComment, "created")

		// when
		err := phony.SendHook(testServer.URL, string(github.IssueComment), mashal(event), secret)

		// then
		Ω(err).ShouldNot(HaveOccurred())
		counter, err := server.HandledEventsCounterWithLabelValues(string(github.IssueComment))
		Ω(err).ShouldNot(HaveOccurred())

		verifyCount(counter, 1)
	})

	It("should get Rate limit for GitHub API calls", func() {
		// given
		setRateLimitMocks()
		event := MockPr().
			LoadedFrom("../plugin/work-in-progress/test_fixtures/github_calls/prs/pr_details.json").
			Create().
			CreateCommentEvent(SentByRepoOwner, testkeeper.BypassCheckComment, "created")

		// when
		err := phony.SendHook(testServer.URL, string(github.IssueComment), mashal(event), secret)

		// then
		Ω(err).ShouldNot(HaveOccurred())

		gauge, err := server.RateLimitWithLabelValues("core")
		Ω(err).ShouldNot(HaveOccurred())

		verifyGauge(gauge, 8)

		gauge, err = server.RateLimitWithLabelValues("search")
		Ω(err).ShouldNot(HaveOccurred())

		verifyGauge(gauge, 10)
	})
})

func mashal(event interface{}) []byte {
	payload, err := json.Marshal(event)
	Ω(err).ShouldNot(HaveOccurred())
	return payload
}

func setRateLimitMocks() {
	gock.New("https://api.github.com").
		Get("/rate_limit").
		Reply(200).
		Body(FromFile("../github/client/test_fixtures/gh/low_rate_limit.json"))
	gock.New("http://127.0.0.1").
		Post("").
		EnableNetworking()
}

func verifyCount(c prometheus.Counter, expected int) {
	count, err := utils.Count(c)
	Ω(err).ShouldNot(HaveOccurred())
	Expect(count).To(Equal(expected))
}

func verifyGauge(g prometheus.Gauge, expected int) {
	gaugeValueSearch, err := utils.GaugeValue(g)
	Ω(err).ShouldNot(HaveOccurred())
	Expect(gaugeValueSearch).To(Equal(expected))
}

package testkeeper_test

import (
	"errors"
	"fmt"
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	gogh "github.com/google/go-github/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("TestKeeper Metrics", func() {

	var handler *testkeeper.GitHubTestEventsHandler
	log := log.NewTestLogger()
	toBe := func(status, description, context, detailsLink string) SoftMatcher {
		return SoftlySatisfyAll(
			HaveState(status),
			HaveDescription(description),
			HaveContext(context),
			HaveTargetURL(fmt.Sprintf("%s/%s/%s.html", docStatusRoot, strings.ToLower(status), detailsLink)),
		)
	}

	toHaveBodyWithWholePluginsComment := SoftlySatisfyAll(
		HaveBodyThatContains(fmt.Sprintf(ghservice.PluginTitleTemplate, testkeeper.ProwPluginName)),
		HaveBodyThatContains("@bartoszmajsak"),
	)

	BeforeEach(func() {
		defer gock.OffAll()
		handler = &testkeeper.GitHubTestEventsHandler{Client: NewDefaultGitHubClient(), BotName: botName}
		testkeeper.UnRegisterAndResetMetrics()
		testkeeper.RegisterMetrics()
	})

	AfterEach(func() {
		testkeeper.UnRegisterAndResetMetrics()
		EnsureGockRequestsHaveBeenMatched()
	})

	It("should report pull requests with /ok-without-tests in histogram", func() {
		//given
		expectedBound := []float64{1, 3, 9, 27, 81, 243}
		expectedCnt := []uint64{0, 1, 1, 2, 2, 2, 2}

		NonExistingRawGitHubFiles("test-keeper.yml", "test-keeper.yaml")

		gock.New("https://api.github.com").
			Get("/repos/" + repositoryName + "/pulls/1").
			Reply(200).
			Body(FromFile("test_fixtures/github_calls/prs/without_tests/pr_details.json"))

		gock.New("https://api.github.com").
			Times(2).
			Get("/repos/" + repositoryName + "/collaborators/bartoszmajsak/permission").
			Reply(200).
			BodyString(`{"permission": "admin"}`)

		gock.New("https://api.github.com").
			Times(2).
			Get("/repos/" + repositoryName + "/pulls/1/reviews").
			Reply(200).
			BodyString(`[]`)

		toHaveEnforcedSuccessState := SoftlySatisfyAll(
			HaveState(github.StatusSuccess),
			HaveDescription(fmt.Sprintf(testkeeper.ApprovedByMessage, "bartoszmajsak")),
		)

		gock.New("https://api.github.com").
			Times(2).
			Post("/repos/" + repositoryName + "/statuses").
			SetMatcher(ExpectPayload(toHaveEnforcedSuccessState)).
			Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

		statusPayload := LoadFromFile("test_fixtures/github_calls/prs/without_tests/skip_comment_by_admin.json")
		issueCommentEvent := TriggerIssueCommentEvent(statusPayload, gogh.IssueCommentEvent{})

		// when
		err := handler.HandleIssueCommentEvent(log, issueCommentEvent)

		// then
		Ω(err).ShouldNot(HaveOccurred())

		// given
		gock.New("https://api.github.com").
			Get("/repos/" + repositoryName + "/pulls/1").
			Reply(200).
			Body(FromFile("test_fixtures/github_calls/prs/without_tests/pr_details_for_metrics.json"))

		statusPayload = LoadFromFile("test_fixtures/github_calls/prs/without_tests/skip_comment_by_admin.json")
		issueCommentEvent = TriggerIssueCommentEvent(statusPayload, gogh.IssueCommentEvent{})

		// when
		err = handler.HandleIssueCommentEvent(log, issueCommentEvent)

		// then - should not expect any additional request mocking
		Ω(err).ShouldNot(HaveOccurred())
		histogram, err := testkeeper.OkWithoutTestsPullRequestWithLabelValues(repositoryName)
		Ω(err).ShouldNot(HaveOccurred())

		metric, err := toMetric(histogram)
		Ω(err).ShouldNot(HaveOccurred())
		verifyHistogram(metric, 2, expectedBound, expectedCnt)
	})

	It("should report pull requests with tests", func() {
		//given
		NonExistingRawGitHubFiles("test-keeper.yml", "test-keeper.yaml", "test-keeper_hint.md")
		gockEmptyComments(2)

		gock.New("https://api.github.com").
			Get("/repos/" + repositoryName + "/pulls/2/files").
			Reply(200).
			Body(FromFile("test_fixtures/github_calls/prs/with_tests/changes.json"))

		gock.New("https://api.github.com").
			Post("/repos/" + repositoryName + "/statuses").
			SetMatcher(ExpectPayload(toBe(github.StatusSuccess, testkeeper.TestsExistMessage, expectedContext, testkeeper.TestsExistDetailsPageName))).
			Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

		statusPayload := LoadFromFile("test_fixtures/github_calls/prs/with_tests/status_opened.json")
		pullRequestEvent := TriggerPullRequestEvent(statusPayload, gogh.PullRequestEvent{})

		// when
		err := handler.HandlePullRequestEvent(log, pullRequestEvent)

		//then
		Ω(err).ShouldNot(HaveOccurred())
		verifyCounter(testkeeper.WithTests, 1)
		verifyCounter(testkeeper.WithoutTests, 0)
	})

	It("should report pull requests without tests", func() {
		//given
		NonExistingRawGitHubFiles("test-keeper.yml", "test-keeper.yaml", "test-keeper_hint.md")
		gockEmptyComments(1)

		gock.New("https://api.github.com").
			Get("/repos/" + repositoryName + "/pulls/1/files").
			Reply(200).
			Body(FromFile("test_fixtures/github_calls/prs/without_tests/changes.json"))

		// This way we implicitly verify that call happened after `HandleEvent` call
		gock.New("https://api.github.com").
			Post("/repos/" + repositoryName + "/statuses").
			SetMatcher(ExpectPayload(toBe(github.StatusFailure, testkeeper.NoTestsMessage, expectedContext, testkeeper.NoTestsDetailsPageName))).
			Reply(201)
		gock.New("https://api.github.com").
			Post("/repos/" + repositoryName + "/issues/1/comments").
			SetMatcher(ExpectPayload(toHaveBodyWithWholePluginsComment)).
			Reply(201)

		statusPayload := LoadFromFile("test_fixtures/github_calls/prs/without_tests/status_opened.json")
		pullRequestEvent := TriggerPullRequestEvent(statusPayload, gogh.PullRequestEvent{})

		// when
		err := handler.HandlePullRequestEvent(log, pullRequestEvent)

		//then
		Ω(err).ShouldNot(HaveOccurred())
		verifyCounter(testkeeper.WithTests, 0)
		verifyCounter(testkeeper.WithoutTests, 1)
	})

})

func verifyHistogram(m *dto.Metric, expectedCount uint64, expectedBound []float64, expectedCnt []uint64) {
	if expectedCount != m.Histogram.GetSampleCount() {
		Fail(fmt.Sprintf("Histogram count was incorrect, expected: %d, actual: %d",
			expectedCount, m.Histogram.GetSampleCount()))
	}
	for ind, bucket := range m.Histogram.GetBucket() {
		if expectedBound[ind] != *bucket.UpperBound {
			Fail(fmt.Sprintf("Bucket upper bound was incorrect, expected: %f, actual: %f\n",
				expectedBound[ind], *bucket.UpperBound))
		}
		if expectedCnt[ind] != *bucket.CumulativeCount {
			Fail(fmt.Sprintf("Bucket cumulative count was incorrect, expected: %d, actual: %d\n",
				expectedCnt[ind], *bucket.CumulativeCount))
		}
	}
}

func verifyCounter(label string, count int) {
	counter, err := testkeeper.PullRequestCounterWithLabelValues(repositoryName, label)
	Ω(err).ShouldNot(HaveOccurred())
	actualCount, err := utils.Count(counter)
	Ω(err).ShouldNot(HaveOccurred())
	Expect(actualCount).To(Equal(count))
}

func toMetric(counter prometheus.Observer) (*dto.Metric, error) {
	metric := &dto.Metric{}
	histogram, ok := counter.(prometheus.Histogram)
	if !ok {
		return nil, errors.New("failed to convert prometheus.Observer to prometheus.Histogram")
	}
	histogram.Write(metric)
	return metric, nil
}

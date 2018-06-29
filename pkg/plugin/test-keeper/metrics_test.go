package testkeeper_test

import (
	"errors"
	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("TestKeeper Metrics", func() {

	var handler *testkeeper.GitHubTestEventsHandler
	var mocker MockPluginTemplate
	log := log.NewTestLogger()

	BeforeEach(func() {
		defer gock.OffAll()
		handler = &testkeeper.GitHubTestEventsHandler{Client: NewDefaultGitHubClient(), BotName: botName}
		testkeeper.UnRegisterAndResetMetrics()
		testkeeper.RegisterMetrics()
		mocker = NewMockPluginTemplate(testkeeper.ProwPluginName)
	})

	AfterEach(func() {
		testkeeper.UnRegisterAndResetMetrics()
		EnsureGockRequestsHaveBeenMatched()
	})

	It("should report pull requests with /ok-without-tests in histogram", func() {
		//given
		expectedBound := []float64{1, 3, 9, 27, 81, 243}
		expectedCnt := []uint64{0, 1, 1, 2, 2, 2, 2}
		approvedBy := fmt.Sprintf(testkeeper.ApprovedByMessage, "admin")

		mockPrBuilder := mocker.MockPr().LoadedFromDefaultJSON().
			WithSize(2).
			WithoutConfigFiles().
			WithUsers(Admin("admin")).
			WithoutReviews().
			Expecting(Status(ToBe(github.StatusSuccess, approvedBy, testkeeper.ApprovedByDetailsPageName)))
		prMock := mockPrBuilder.Create()

		event := prMock.CreateCommentEvent(SentBy("admin"), testkeeper.BypassCheckComment, "created")

		// when
		err := handler.HandleIssueCommentEvent(log, event)

		// then
		Ω(err).ShouldNot(HaveOccurred())

		// given
		mockPrBuilder.
			WithSize(26).
			Create()

		// when
		err = handler.HandleIssueCommentEvent(log, event)

		// then - should not expect any additional request mocking
		Ω(err).ShouldNot(HaveOccurred())
		repositoryName := *prMock.PullRequest.Base.Repo.FullName
		histogram, err := testkeeper.OkWithoutTestsPullRequestWithLabelValues(repositoryName)
		Ω(err).ShouldNot(HaveOccurred())

		metric, err := toMetric(histogram)
		Ω(err).ShouldNot(HaveOccurred())
		verifyHistogram(metric, 2, expectedBound, expectedCnt)
	})

	It("should report pull requests with tests", func() {
		//given
		prMock := mocker.MockPr().LoadedFromDefaultJSON().
			WithoutConfigFiles().
			WithoutComments().
			WithFiles(LoadedFrom("test_fixtures/github_calls/prs/with_tests/changes.json")).
			Expecting(
				Status(ToBe(github.StatusSuccess, testkeeper.TestsExistMessage, testkeeper.TestsExistDetailsPageName))).
			Create()

		// when
		err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("opened"))

		//then
		Ω(err).ShouldNot(HaveOccurred())
		verifyCounter(testkeeper.WithTests, 1, prMock)
		verifyCounter(testkeeper.WithoutTests, 0, prMock)
	})

	It("should report pull requests without tests", func() {
		//given
		prMock := mocker.MockPr().LoadedFromDefaultJSON().
			WithoutConfigFiles().
			WithoutMessageFiles("test-keeper_without_tests_message.md").
			WithoutComments().
			WithFiles(LoadedFrom("test_fixtures/github_calls/prs/without_tests/changes.json")).
			Expecting(
				Status(ToBe(github.StatusFailure, testkeeper.NoTestsMessage, testkeeper.NoTestsDetailsPageName)),
				Comment(ContainingStatusMessage(testkeeper.WithoutTestsMsg))).
			Create()

		// when
		err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("opened"))

		//then
		Ω(err).ShouldNot(HaveOccurred())
		verifyCounter(testkeeper.WithTests, 0, prMock)
		verifyCounter(testkeeper.WithoutTests, 1, prMock)
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

func verifyCounter(label string, count int, prMock *PrMock) {
	repositoryName := *prMock.PullRequest.Base.Repo.FullName
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

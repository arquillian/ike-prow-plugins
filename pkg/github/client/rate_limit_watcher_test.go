package ghclient_test

import (
	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus/hooks/test"
	"gopkg.in/h2non/gock.v1"
	gogh "github.com/google/go-github/github"
)

var _ = Describe("Rate limit watcher", func() {

	logger, hook := test.NewNullLogger()
	client := ghclient.NewClient(gogh.NewClient(nil), logger)
	client.RegisterAroundFunctions(
		ghclient.NewRateLimitWatcher(client, logger, 10),
		ghclient.NewRetryWrapper(3, 0),
		ghclient.NewPaginationChecker())

	BeforeEach(func() {
		defer gock.OffAll()
		hook.Reset()
	})

	AfterEach(EnsureGockRequestsHaveBeenMatched)

	It("should not log rate limit when within the threshold", func() {
		// given
		mockHighRateLimit()

		gock.New("https://api.github.com").
			Get("/repos/owner/repo/pulls/123/files").
			Reply(200).
			BodyString("[]")

		// when
		_, err := client.ListPullRequestFiles("owner", "repo", 123)

		// then
		Ω(err).ShouldNot(HaveOccurred())
		Expect(hook.Entries).To(BeEmpty())
	})

	It("should log rate limit when within the threshold", func() {
		// given
		gock.New("https://api.github.com").
			Get("/rate_limit").
			Persist().
			Reply(200).
			Body(FromFile("test_fixtures/gh/low_rate_limit.json"))

		gock.New("https://api.github.com").
			Get("/repos/owner/repo/pulls/123/files").
			Reply(200).
			BodyString("[]")

		// when
		_, err := client.ListPullRequestFiles("owner", "repo", 123)

		// then
		Ω(err).ShouldNot(HaveOccurred())
		Expect(hook.Entries).To(HaveLen(1))
		Expect(hook.LastEntry().Message).To(HavePrefix("reaching limit for GH API calls. 8/20 left. resetting at"))
	})
})

func mockHighRateLimit() {
	gock.New("https://api.github.com").
		Get("/rate_limit").
		Persist().
		Reply(200).
		Body(FromFile("test_fixtures/gh/high_rate_limit.json"))
}

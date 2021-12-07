package ghclient_test

import (
	"net/http"

	ghclient "github.com/arquillian/ike-prow-plugins/pkg/github/client"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	gogh "github.com/google/go-github/v41/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gock "gopkg.in/h2non/gock.v1"
)

var _ = Describe("Retry client features", func() {

	client := ghclient.NewClient(gogh.NewClient(nil), log.NewTestLogger())
	client.RegisterAroundFunctions(
		ghclient.NewRateLimitWatcher(client, log.NewTestLogger(), 100),
		ghclient.NewRetryWrapper(3, 0),
		ghclient.NewPaginationChecker())

	Context("Client should try 3 times to get the correct response", func() {

		BeforeEach(func() {
			defer gock.OffAll()
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should try to get the response 3 times and then fail when client gets only 404", func() {
			// given
			calls := 0
			mockHighRateLimit()
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/pulls/123/files").
				SetMatcher(spyOnCalls(&calls)).
				Persist().
				Reply(404).
				BodyString("Not Found")

			// when
			_, err := client.ListPullRequestFiles("owner", "repo", 123)

			// then
			Ω(err).Should(HaveOccurred())
			Expect(calls).To(Equal(3))
		})

		It("should stop resending requests and not fail when client gets 408 and then 200", func() {
			// given
			calls := 0
			mockHighRateLimit()
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/pulls/123/files").
				SetMatcher(spyOnCalls(&calls)).
				Reply(408).
				BodyString("Request Timeout")

			gock.New("https://api.github.com").
				Get("/repos/owner/repo/pulls/123/files").
				SetMatcher(spyOnCalls(&calls)).
				Reply(200).
				BodyString("[]")

			// when
			_, err := client.ListPullRequestFiles("owner", "repo", 123)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(calls).To(Equal(2))
		})
	})

})

func spyOnCalls(counter *int) gock.Matcher {
	matcher := gock.NewBasicMatcher()
	matcher.Add(func(_ *http.Request, _ *gock.Request) (bool, error) { // nolint:unparam
		*counter++
		return true, nil
	})
	return matcher
}

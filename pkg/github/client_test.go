package github_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
)

var _ = Describe("Client features", func() {

	client := NewDefaultGitHubClient()

	Context("Client should try 3 times to get the correct response", func() {

		BeforeEach(func() {
			defer gock.OffAll()
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should try to get the response 3 times and then fail when client gets only 404", func() {
			// given
			counter := 0

			gock.New("https://api.github.com").
				Get("/repos/owner/repo/pulls/123/files").
				SetMatcher(createCounterMather(&counter)).
				Persist().
				Reply(404).
				BodyString("Not Found")

			// when
			_, err := client.ListPullRequestFiles("owner", "repo", 123)

			// then
			Ω(err).Should(HaveOccurred())
			Expect(counter).To(Equal(3))
		})

		It("should stop resending requests and not fail when client gets 408 and then 200", func() {
			// given
			counter := 0

			gock.New("https://api.github.com").
				Get("/repos/owner/repo/pulls/123/files").
				SetMatcher(createCounterMather(&counter)).
				Reply(408).
				BodyString("Request Timeout")

			gock.New("https://api.github.com").
				Get("/repos/owner/repo/pulls/123/files").
				SetMatcher(createCounterMather(&counter)).
				Reply(200).
				BodyString("[]")

			// when
			_, err := client.ListPullRequestFiles("owner", "repo", 123)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(counter).To(Equal(2))
		})
	})

})

func createCounterMather(counter *int) gock.Matcher {
	matcher := gock.NewBasicMatcher()
	matcher.Add(func(_ *http.Request, _ *gock.Request) (bool, error) {
		*counter++
		return true, nil
	})
	return matcher
}

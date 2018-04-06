package github_test

import (
	"net/http"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("Client features", func() {

	client := NewDefaultGitHubClient()

	Context("Client should try 3 times to get the correct response", func() {

		BeforeEach(func() {
			gock.Off()
		})

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

	Context("Lazy loading feature", func() {

		BeforeEach(func() {
			gock.Off()
		})

		It("should load pull request when load() method is called", func() {
			// given
			counter := 0
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/pulls/123").
				SetMatcher(createCounterMather(&counter)).
				Reply(200).
				BodyString(`{"title":"Loaded PR"}`)
			loader := &github.PullRequestLoader{Client: client, RepoOwner: "owner", RepoName: "repo", Number: 123}
			Expect(counter).To(Equal(0))

			// when
			pullRequest, err := loader.Load()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(counter).To(Equal(1))
			Expect(*pullRequest.Title).To(Equal("Loaded PR"))
		})

		It("should load pull request only once", func() {
			// given
			counter := 0
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/pulls/123").
				SetMatcher(createCounterMather(&counter)).
				Persist().
				Reply(200).
				BodyString(`{"title":"Loaded PR"}`)
			loader := &github.PullRequestLoader{Client: client, RepoOwner: "owner", RepoName: "repo", Number: 123}
			loader.Load()

			// when
			pullRequest, err := loader.Load()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(counter).To(Equal(1))
			Expect(*pullRequest.Title).To(Equal("Loaded PR"))
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

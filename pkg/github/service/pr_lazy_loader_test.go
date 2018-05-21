package ghservice_test

import (
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("Pull Request lazy loading", func() {

	client := NewDefaultGitHubClient()

	BeforeEach(func() {
		defer gock.OffAll()
	})

	AfterEach(EnsureGockRequestsHaveBeenMatched)

	It("should load pull request when load() method is called", func() {
		// given
		calls := 0
		gock.New("https://api.github.com").
			Get("/repos/owner/repo/pulls/123").
			SetMatcher(SpyOnCalls(&calls)).
			Reply(200).
			BodyString(`{"title":"Loaded PR"}`)
		loader := &ghservice.PullRequestLazyLoader{Client: client, RepoOwner: "owner", RepoName: "repo", Number: 123}
		Expect(calls).To(Equal(0))

		// when
		pullRequest, err := loader.Load()

		// then
		Ω(err).ShouldNot(HaveOccurred())
		Expect(calls).To(Equal(1))
		Expect(*pullRequest.Title).To(Equal("Loaded PR"))
	})

	It("should load pull request only once", func() {
		// given
		counter := 0
		gock.New("https://api.github.com").
			Get("/repos/owner/repo/pulls/123").
			SetMatcher(SpyOnCalls(&counter)).
			Persist().
			Reply(200).
			BodyString(`{"title":"Loaded PR"}`)
		loader := &ghservice.PullRequestLazyLoader{Client: client, RepoOwner: "owner", RepoName: "repo", Number: 123}
		loader.Load()

		// when
		pullRequest, err := loader.Load()

		// then
		Ω(err).ShouldNot(HaveOccurred())
		Expect(counter).To(Equal(1))
		Expect(*pullRequest.Title).To(Equal("Loaded PR"))
	})
})

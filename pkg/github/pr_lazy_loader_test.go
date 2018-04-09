package github_test

import (
	"github.com/arquillian/ike-prow-plugins/pkg/github"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("Pull Request lazy loading", func() {

	client := NewDefaultGitHubClient()

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

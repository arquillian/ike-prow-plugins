package github_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
	"github.com/arquillian/ike-prow-plugins/plugin/scm"
	gogh "github.com/google/go-github/github"
	"github.com/arquillian/ike-prow-plugins/plugin/github"
	"github.com/arquillian/ike-prow-plugins/plugin/internal/test"
)

var _ = Describe("Repository Service", func() {

	Context("Languages used in the repository", func() {

		BeforeEach(func() {
			defer gock.Off()
		})

		It("should get all languages used in repository", func() {
			// given
			languageResponse :=
				`{
					"Go": 48810,
					"Makefile": 4420,
					"Shell": 1527,
					"Ruby": 226
				}`
			repositoryService := github.RepositoryService{
				Client: gogh.NewClient(nil),
				Change: scm.RepositoryChange{
					Owner:    "arquillian",
					RepoName: "ike-prow-plugins",
					Hash:     "9c483d7bd570eed80d05f27e81d45147dcf68869",
				},
				Log: test.CreateNullLogger(),
			}

			gock.New("https://api.github.com/").
				Get("/repos/arquillian/ike-prow-plugins/languages").
				Reply(200).
				BodyString(languageResponse)

			// when
			usedLanguages, err := repositoryService.UsedLanguages()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(usedLanguages).To(ConsistOf("Go", "Shell", "Ruby", "Makefile"))
		})

		It("should return empty slice when no languages found in the repository", func() {
			// given
			repositoryService := github.RepositoryService{
				Client: gogh.NewClient(nil),
				Change: scm.RepositoryChange{
					Owner:    "arquillian",
					RepoName: "ike-prow-plugins",
					Hash:     "9c483d7bd570eed80d05f27e81d45147dcf68869",
				},
				Log: test.CreateNullLogger(),
			}

			gock.New("https://api.github.com/").
				Get("/repos/arquillian/ike-prow-plugins/languages").
				Reply(200).
				BodyString("{}")

			// when
			usedLanguages, err := repositoryService.UsedLanguages()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(usedLanguages).To(BeEmpty())

		})
	})
})

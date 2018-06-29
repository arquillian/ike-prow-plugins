package wip_test

import (
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/work-in-progress"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("Work In Progress config loader features", func() {

	var mocker = NewMockPluginTemplate(wip.ProwPluginName)

	BeforeEach(func() {
		defer gock.OffAll()
	})

	AfterEach(EnsureGockRequestsHaveBeenMatched)

	Context("Loading work-in-progress configuration file from GitHub repository", func() {

		logger := log.NewTestLogger()

		It("should load work-in-progress configuration yml file", func() {
			// given
			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			mocker.AddConfig(
				ConfigYml(Containing(
					Param("title_prefixes", "['[work in progress]', 'work in progress']"),
					Param("gh_label", "working-in-progress")))).
				ToChange(change)

			// when
			configuration := wip.LoadConfiguration(logger, change)

			// then
			Expect(configuration.Prefix).To(ConsistOf("[work in progress]", "work in progress"))
			Expect(configuration.Combine).To(Equal(true))
			Expect(configuration.Label).To(Equal("working-in-progress"))
		})

		It("should not load work-in-progress configuration yaml file and return empty url when config is not accessible", func() {
			// given
			NonExistingRawGitHubFiles("work-in-progress.yml", "work-in-progress.yaml")

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			// when
			configuration := wip.LoadConfiguration(logger, change)

			// then
			Expect(configuration.Prefix).To(BeEmpty())
			Expect(configuration.Combine).To(Equal(true))
			Expect(configuration.Label).To(Equal("work-in-progress"))
		})
	})
})

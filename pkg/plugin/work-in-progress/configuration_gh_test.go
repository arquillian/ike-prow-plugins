package wip_test

import (
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/work-in-progress"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
)

var _ = Describe("Work In Progress config loader features", func() {

	BeforeEach(func() {
		defer gock.OffAll()
	})

	AfterEach(EnsureGockRequestsHaveBeenMatched)

	Context("Loading test-keeper configuration file from GitHub repository", func() {

		logger := log.NewTestLogger()
		configFilePath := ghservice.ConfigHome + wip.ProwPluginName

		It("should load work-in-progress configuration yml file", func() {
			// given
			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/" + configFilePath + ".yml").
				Reply(200).
				BodyString("title_prefixes: ['[work in progress]', 'work in progress']")

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			// when
			configuration := wip.LoadConfiguration(logger, change)

			// then
			Expect(configuration.Prefix).To(ConsistOf("[work in progress]", "work in progress"))
			Expect(configuration.Combine).To(Equal(true))
		})

		It("should not load work-in-progress configuration yaml file and return empty url when config is not accessible", func() {
			// given
			NonExistingRawGitHubFiles("work-in-progress.yml")

			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/" + configFilePath + ".yaml").
				Reply(404)

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
		})
	})
})

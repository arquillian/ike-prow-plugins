package prsanitizer_test

import (
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/pr-sanitizer"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("PR Sanitizer config loader features", func() {

	BeforeEach(func() {
		defer gock.OffAll()
	})

	AfterEach(EnsureGockRequestsHaveBeenMatched)

	Context("Loading pr-sanitizer configuration file from GitHub repository", func() {

		logger := log.NewTestLogger()
		configFilePath := ghservice.ConfigHome + prsanitizer.ProwPluginName

		It("should load pr-sanitizer configuration yml file", func() {
			// given
			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/" + configFilePath + ".yml").
				Reply(200).
				BodyString("type_prefixes: [':star:', ':package:', ':hammer_and_wrench:']\n" +
				"description_content_length: 40")

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			// when
			configuration := prsanitizer.LoadConfiguration(logger, change)

			// then
			Expect(configuration.TypePrefix).To(ConsistOf(":star:", ":package:", ":hammer_and_wrench:"))
			Expect(configuration.Combine).To(Equal(true))
			Expect(configuration.DescriptionContentLength).To(Equal(40))
		})

		It("should not load pr-sanitizer configuration yaml file and return empty url when config is not accessible", func() {
			// given
			NonExistingRawGitHubFiles("pr-sanitizer.yml", "pr-sanitizer.yaml")

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			// when
			configuration := prsanitizer.LoadConfiguration(logger, change)

			// then
			Expect(configuration.TypePrefix).To(BeEmpty())
			Expect(configuration.Combine).To(Equal(true))
		})
	})
})

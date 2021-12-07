package testkeeper_test

import (
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	testkeeper "github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gock "gopkg.in/h2non/gock.v1"
)

var _ = Describe("Test keeper config loader features", func() {

	var mocker = NewMockPluginTemplate(testkeeper.ProwPluginName)

	BeforeEach(func() {
		defer gock.OffAll()
	})

	AfterEach(EnsureGockRequestsHaveBeenMatched)

	Context("Loading test-keeper configuration file from GitHub repository", func() {

		logger := log.NewTestLogger()

		It("should load test-keeper configuration yml file", func() {
			// given
			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			mocker.AddConfig(
				ConfigYml(Containing(
					Param("test_patterns", "['*my', 'test.go', 'pattern.js']"),
					Param("skip_validation_for", "['pom.xml', 'regex{{*\\.adoc}}']")))).
				ToChange(change)

			// when
			configuration := testkeeper.LoadConfiguration(logger, change)

			// then
			Expect(configuration.LocationURL).To(Equal("https://github.com/owner/repo/blob/46cb8fac44709e4ccaae97448c65e8f7320cfea7/.ike-prow/test-keeper.yml"))
			Expect(configuration.PluginName).To(Equal(testkeeper.ProwPluginName))
			Expect(configuration.Inclusions).To(ConsistOf("*my", "test.go", "pattern.js"))
			Expect(configuration.Exclusions).To(ConsistOf("pom.xml", "regex{{*\\.adoc}}"))
			Expect(configuration.Combine).To(BeTrue())
		})

		It("should load test-keeper configuration yaml file", func() {
			// given
			NonExistingRawGitHubFiles(".ike-prow/test-keeper.yml")

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			mocker.AddConfig(
				ConfigYaml(Containing(
					Param("test_patterns", "['*my', 'test.go', 'pattern.js']"),
					Param("skip_validation_for", "['pom.xml', 'regex{{*\\.adoc}}']")))).
				ToChange(change)

			// when
			configuration := testkeeper.LoadConfiguration(logger, change)

			// then
			Expect(configuration.Inclusions).To(ConsistOf("*my", "test.go", "pattern.js"))
			Expect(configuration.Exclusions).To(ConsistOf("pom.xml", "regex{{*\\.adoc}}"))
			Expect(configuration.Combine).To(BeTrue())
		})

		It("should not load test-keeper configuration yaml file and return empty url when config is not accessible", func() {
			// given
			NonExistingRawGitHubFiles(".ike-prow/test-keeper.yml", ".ike-prow/test-keeper.yaml")

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			// when
			configuration := testkeeper.LoadConfiguration(logger, change)

			// then
			Expect(configuration.LocationURL).To(BeEmpty())
			Expect(configuration.Inclusions).To(BeEmpty())
			Expect(configuration.Exclusions).To(BeEmpty())
			Expect(configuration.Combine).To(BeTrue())
		})
	})
})

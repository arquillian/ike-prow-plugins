package testkeeper_test

import (
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("Test keeper config loader features", func() {

	BeforeEach(func() {
		defer gock.OffAll()
	})

	AfterEach(EnsureGockRequestsHaveBeenMatched)

	Context("Loading test-keeper configuration file from GitHub repository", func() {

		logger := log.NewTestLogger()

		It("should load test-keeper configuration yml file", func() {
			// given
			NonExistingRawGitHubFiles(".ike-prow/test-keeper_hint.md")

			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/.ike-prow/" + testkeeper.ProwPluginName + ".yml").
				Reply(200).
				BodyString("test_patterns: ['*my', 'test.go', 'pattern.js']\n" +
					"skip_validation_for: ['pom.xml', 'regex{{*\\.adoc}}']\n" +
					"plugin_hint: 'http://my.server.com/message.md'")

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			// when
			configuration := testkeeper.LoadConfiguration(logger, change)

			// then
			Expect(configuration.LocationURL).To(Equal("https://github.com/owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/.ike-prow/test-keeper.yml"))
			Expect(configuration.PluginName).To(Equal(testkeeper.ProwPluginName))
			Expect(configuration.PluginHint).To(Equal("http://my.server.com/message.md"))
		})

		It("should load test-keeper configuration yml file", func() {
			// given
			NonExistingRawGitHubFiles(".ike-prow/test-keeper_hint.md")

			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/.ike-prow/" + testkeeper.ProwPluginName + ".yml").
				Reply(200).
				BodyString("test_patterns: ['*my', 'test.go', 'pattern.js']\n" +
					"skip_validation_for: ['pom.xml', 'regex{{*\\.adoc}}']\n" +
					"plugin_hint: 'http://my.server.com/message.md'")

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			// when
			configuration := testkeeper.LoadConfiguration(logger, change)

			// then
			Expect(configuration.PluginHint).To(Equal("http://my.server.com/message.md"))
			Expect(configuration.Inclusions).To(ConsistOf("*my", "test.go", "pattern.js"))
			Expect(configuration.Exclusions).To(ConsistOf("pom.xml", "regex{{*\\.adoc}}"))
			Expect(configuration.Combine).To(BeTrue())
		})

		It("should load test-keeper configuration yaml file", func() {
			// given
			NonExistingRawGitHubFiles(".ike-prow/test-keeper.yml")
			NonExistingRawGitHubFiles(".ike-prow/test-keeper_hint.md")

			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/.ike-prow/" + testkeeper.ProwPluginName + ".yaml").
				Reply(200).
				BodyString("test_patterns: ['*my', 'test.go', 'pattern.js']\n" +
					"skip_validation_for: ['pom.xml', 'regex{{*\\.adoc}}']\n" +
					"plugin_hint: 'http://my.server.com/message.md'")

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			// when
			configuration := testkeeper.LoadConfiguration(logger, change)

			// then
			Expect(configuration.PluginHint).To(Equal("http://my.server.com/message.md"))
			Expect(configuration.Inclusions).To(ConsistOf("*my", "test.go", "pattern.js"))
			Expect(configuration.Exclusions).To(ConsistOf("pom.xml", "regex{{*\\.adoc}}"))
			Expect(configuration.Combine).To(BeTrue())
		})

		It("should load test-keeper configuration yaml file with default hint file if present and not mentioned in configuration", func() {
			// given
			NonExistingRawGitHubFiles(".ike-prow/test-keeper.yml")

			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/.ike-prow/" + testkeeper.ProwPluginName + ".yaml").
				Reply(200).
				BodyString("test_patterns: ['*my', 'test.go', 'pattern.js']\n" +
				"skip_validation_for: ['pom.xml', 'regex{{*\\.adoc}}']\n")

			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/.ike-prow/" + testkeeper.ProwPluginName + "_hint.md").
				Reply(200).
				BodyString("custom hint for plugin")

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			// when
			configuration := testkeeper.LoadConfiguration(logger, change)

			// then
			Expect(configuration.PluginHint).To(Equal("https://github.com/owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/.ike-prow/test-keeper_hint.md"))
			Expect(configuration.Inclusions).To(ConsistOf("*my", "test.go", "pattern.js"))
			Expect(configuration.Exclusions).To(ConsistOf("pom.xml", "regex{{*\\.adoc}}"))
			Expect(configuration.Combine).To(BeTrue())
		})

		It("should not load test-keeper configuration yaml file and return empty url when config is not accessible", func() {
			// given
			NonExistingRawGitHubFiles(".ike-prow/test-keeper.yml")

			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/.ike-prow/" + testkeeper.ProwPluginName + ".yaml").
				Reply(404)

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			// when
			configuration := testkeeper.LoadConfiguration(logger, change)

			// then
			Expect(configuration.LocationURL).To(BeEmpty())
			Expect(configuration.PluginHint).To(BeEmpty())
			Expect(configuration.Inclusions).To(BeEmpty())
			Expect(configuration.Exclusions).To(BeEmpty())
			Expect(configuration.Combine).To(BeTrue())
		})
	})
})

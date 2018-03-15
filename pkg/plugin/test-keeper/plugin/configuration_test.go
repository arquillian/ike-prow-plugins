package plugin_test

import (
	"github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper/plugin"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("Test keeper config loader features", func() {

	BeforeEach(func() {
		gock.Off()
	})

	Context("Loading test-keeper configuration file from the repository", func() {

		It("should load test-keeper configuration yaml file", func() {
			// given
			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/" + plugin.ProwPluginName + ".yml").
				Reply(200).
				BodyString("test_pattern: (.*my|test\\.go|pattern\\.js)$\n" +
					"skip_validation_for: pom\\.xml|*\\.adoc\n" +
					"plugin_hint: 'http://my.server.com/message.md'")

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			// when
			configuration := plugin.LoadTestKeeperConfig(test.CreateNullLogger(), change)

			// then
			Expect(configuration.LocationURL).NotTo(Equal(""))
			Expect(configuration.Inclusion).To(Equal(`(.*my|test\.go|pattern\.js)$`))
			Expect(configuration.Exclusion).To(Equal(`pom\.xml|*\.adoc`))
			Expect(configuration.Combine).To(BeTrue())
			Expect(configuration.PluginHint).To(Equal("http://my.server.com/message.md"))
		})

		It("should not load test-keeper configuration yaml file and return empty url when config is not accessible", func() {
			// given
			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			// when
			configuration := plugin.LoadTestKeeperConfig(test.CreateNullLogger(), change)

			// then
			Expect(configuration.LocationURL).To(Equal(""))
			Expect(configuration.Inclusion).To(Equal(""))
			Expect(configuration.Exclusion).To(Equal(""))
			Expect(configuration.Combine).To(BeTrue())
			Expect(configuration.PluginHint).To(Equal(""))
		})
	})
})

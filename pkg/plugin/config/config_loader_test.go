package config_test

import (
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/config"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gock "gopkg.in/h2non/gock.v1"
)

type sampleConfiguration struct {
	Inclusion string `yaml:"test_pattern,omitempty"`
	Exclusion string `yaml:"exclusion,omitempty"`
	Combine   bool   `yaml:"combine_defaults,omitempty"`
	AnyNumber int    `yaml:"number,omitempty"`
}

var _ = Describe("Config loader features", func() {

	BeforeEach(func() {
		gock.Off()
	})

	Context("Loading configuration file from the repository", func() {

		It("should load configuration yaml file", func() {
			// given
			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/sample-plugin.yml").
				Reply(200).
				BodyString("test_pattern: (.*my|test\\.go|pattern\\.js)$\n" +
					"exclusion: pom\\.xml|*\\.adoc\n" +
					"number: 12345")

			loader := config.NewPluginConfigLoader("sample-plugin",
				scm.RepositoryChange{
					Owner:    "owner",
					RepoName: "repo",
					Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
				})

			configuration := sampleConfiguration{
				Combine: true,
			}

			// when
			err := loader.Load(&configuration)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(configuration.Inclusion).To(Equal(`(.*my|test\.go|pattern\.js)$`))
			Expect(configuration.Exclusion).To(Equal(`pom\.xml|*\.adoc`))
			Expect(configuration.Combine).To(BeTrue())
			Expect(configuration.AnyNumber).To(Equal(12345))
		})
	})

	Context("URL construction", func() {

		It("should create a URL to a configuration file for the given change", func() {
			// given
			loader := config.NewPluginConfigLoader("sample-plugin",
				scm.RepositoryChange{
					Owner:    "owner",
					RepoName: "repo",
					Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
				})

			// when
			url := loader.CreateConfigFileURL()

			// then
			Expect(url).To(Equal("https://raw.githubusercontent.com/" +
				"owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/sample-plugin.yml"))
		})
	})
})

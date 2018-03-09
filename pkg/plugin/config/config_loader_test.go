package config

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
)

type sampleConfiguration struct {
	Inclusion string `yaml:"test_pattern,omitempty"`
}

var _ = Describe("Config loader features", func() {

	BeforeEach(func() {
		gock.Off()
	})

	Context("Loading configuration file from the repository", func() {

		It("should load configuration yaml file", func() {
			// given
			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/sample-config.yml").
				Reply(200).
				BodyString(`test_pattern: (.*my|test\.go|pattern\.js)$`)

			loader := PluginConfigLoader{
				PluginName: "sample-config",
				Change: scm.RepositoryChange{
					Owner:    "owner",
					RepoName: "repo",
					Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
				},
			}

			configuration := sampleConfiguration{}

			// when
			err := loader.Load(&configuration)
			
			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(configuration.Inclusion).To(Equal(`(.*my|test\.go|pattern\.js)$`))
		})
	})
})
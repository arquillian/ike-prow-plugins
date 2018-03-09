package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/arquillian/ike-prow-plugins/plugin/internal/test"
	"gopkg.in/h2non/gock.v1"
	"github.com/arquillian/ike-prow-plugins/plugin/config"
	"github.com/arquillian/ike-prow-plugins/plugin/test-keeper/plugin"
	"github.com/arquillian/ike-prow-plugins/plugin/scm"
)

var _ = Describe("Config loader features", func() {

	BeforeEach(func() {
		gock.Off()
	})

	Context("Loading configuration file from the repository", func() {

		It("should load test-keeper configuration", func() {
			// given
			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/test-keeper.yml").
				Reply(200).
				Body(FromFile("test_fixtures/test-keeper.yml"))

			loader := config.PluginConfigLoader{
				PluginName: "test-keeper",
				Change: scm.RepositoryChange{
					Owner:    "owner",
					RepoName: "repo",
					Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
				},
			}

			configuration := plugin.TestKeeperConfiguration{}

			// when
			err := loader.Load(&configuration)
			
			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(configuration.Inclusion).To(Equal(".*my|test.go|pattern.js"))
		})

	})
})
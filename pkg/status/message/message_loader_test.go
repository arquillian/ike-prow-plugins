package message_test

import (
	"github.com/arquillian/ike-prow-plugins/pkg/config"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/status/message"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

const ProwPluginName string = "my-test-plugin"

var _ = Describe("Loader message creation", func() {

	Context("Creation of default message messages that are sent to a validated PR when custom message file is not set", func() {

		It("should create default message referencing to documentation when url to config is empty", func() {
			// given
			conf := config.PluginConfiguration{PluginName: ProwPluginName}
			message := &message.Loader{
				Message: &message.Message{ConfigFile: conf.LocationURL, Documentation: "#_my_test_plugin"},
			}

			// when
			msg := message.LoadMessage(scm.RepositoryChange{}, "")

			// then
			Expect(msg).To(ContainSubstring("http://arquillian.org/ike-prow-plugins/#_my_test_plugin"))
			Expect(msg).NotTo(ContainSubstring("Your plugin configuration is stored in"))
		})

		It("should create default message referencing to config file when url to config is not empty", func() {
			// given
			url := "http://github.com/my/repo/my-test-plugin.yaml"
			conf := config.PluginConfiguration{LocationURL: url}
			message := &message.Loader{
				Message: &message.Message{ConfigFile: conf.LocationURL, Documentation: "#_my_test_plugin"},
			}

			// when
			msg := message.LoadMessage(scm.RepositoryChange{}, "")

			// then
			Expect(msg).NotTo(ContainSubstring("http://arquillian.org/ike-prow-plugins/#_my_test_plugin"))
			Expect(msg).To(ContainSubstring(url))
		})
	})

	Context("Creation of default message messages from default location when config plugin message is not set", func() {

		BeforeEach(func() {
			defer gock.OffAll()
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should create message taken from a default message file if plugin message url is missing", func() {
			// given
			NonExistingRawGitHubFiles("my-test-plugin.yaml", "my-test-plugin.yml")

			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/.ike-prow/my-test-plugin_message.md").
				Reply(200).
				BodyString("Custom message")

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			message := &message.Loader{
				PluginName: ProwPluginName,
				Message: &message.Message{
					ConfigFile:    "http://github.com/my/repo/my-test-plugin.yaml",
					Documentation: "#_my_test_plugin"},
			}

			// when
			msg := message.LoadMessage(change, "")

			// then
			Expect(msg).To(Equal("Custom message"))
		})

		It("should create message taken from a default message file if location url is missing", func() {
			// given
			NonExistingRawGitHubFiles("my-test-plugin.yaml", "my-test-plugin.yml")

			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/.ike-prow/my-test-plugin_message.md").
				Reply(200).
				BodyString("Custom message")

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			message := &message.Loader{
				PluginName: ProwPluginName,
				Message:    &message.Message{Documentation: "#_my_test_plugin"},
			}

			// when
			msg := message.LoadMessage(change, "")

			// then
			Expect(msg).To(Equal("Custom message"))
		})
	})
})

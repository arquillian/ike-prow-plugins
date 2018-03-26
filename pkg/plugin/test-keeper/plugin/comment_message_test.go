package plugin_test

import (
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/plugin/config"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper/plugin"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/microcosm-cc/bluemonday"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("Test keeper comment message creation", func() {

	Context("Creation of default comment messages that are sent to a validated PR when custom message file is not set", func() {

		It("should create default message referencing to documentation when url to config is empty", func() {
			// given
			config := plugin.TestKeeperConfiguration{PluginConfiguration: config.PluginConfiguration{PluginHint: "any-file"}}

			// when
			msg := plugin.CreateCommentMessage(config, scm.RepositoryChange{})

			// then
			Expect(msg).To(ContainSubstring("http://arquillian.org/ike-prow-plugins/#_test_keeper_plugin"))
			Expect(msg).To(ContainSubstring(plugin.SkipComment))
		})

		It("should create default message referencing to config file when url to config is not empty", func() {
			// given
			url := "http://github.com/my/repo/test-keeper.yaml"
			config := plugin.TestKeeperConfiguration{PluginConfiguration: config.PluginConfiguration{LocationURL: url}}

			// when
			msg := plugin.CreateCommentMessage(config, scm.RepositoryChange{})

			// then
			Expect(msg).NotTo(ContainSubstring("http://arquillian.org/ike-prow-plugins/#_test_keeper_plugin"))
			Expect(msg).To(ContainSubstring(url))
			Expect(msg).To(ContainSubstring(plugin.SkipComment))
		})
	})

	Context("Creation of default comment messages that are sent to a validated PR when custom message file is set", func() {

		BeforeEach(func() {
			gock.Off()
		})

		It("should create message taken from a file set in config using relative path", func() {
			// given
			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/path/to/custom_message_file.md").
				Reply(200).
				BodyString("Custom message")

			url := "http://github.com/my/repo/test-keeper.yaml"
			config := plugin.TestKeeperConfiguration{
				PluginConfiguration: config.PluginConfiguration{
					LocationURL: url,
					PluginHint:  "path/to/custom_message_file.md",
				},
			}

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			// when
			msg := plugin.CreateCommentMessage(config, change)
			sanitizedMsg := removeHtmlElements(msg)

			// then
			Expect(string(sanitizedMsg)).To(Equal("Custom message"))
		})

		It("should create default message with no-found-custom-file suffix using wrong relative path", func() {
			// given
			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/path/to/custom_message_file.md").
				Reply(404)

			url := "http://github.com/my/repo/test-keeper.yaml"
			config := plugin.TestKeeperConfiguration{
				PluginConfiguration: config.PluginConfiguration{
					LocationURL: url,
					PluginHint:  "path/to/custom_message_file.md",
				},
			}

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			// when
			msg := plugin.CreateCommentMessage(config, change)

			// then
			Expect(msg).NotTo(ContainSubstring("http://arquillian.org/ike-prow-plugins/#_test_keeper_plugin"))
			Expect(msg).To(ContainSubstring(url))
			Expect(msg).To(ContainSubstring(plugin.SkipComment))
			Expect(msg).To(ContainSubstring(
				"https://raw.githubusercontent.com/owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/" +
					"path/to/custom_message_file.md"))
		})

		It("should create message taken from a file set in config using url", func() {
			// given
			gock.New("http://my.server.com").
				Get("path/to/custom_message_file.md").
				Reply(200).
				BodyString("Custom message")

			url := "http://github.com/my/repo/test-keeper.yaml"
			config := plugin.TestKeeperConfiguration{
				PluginConfiguration: config.PluginConfiguration{
					LocationURL: url,
					PluginHint:  "http://my.server.com/path/to/custom_message_file.md",
				},
			}

			// when
			msg := plugin.CreateCommentMessage(config, scm.RepositoryChange{})
			sanitizedMsg := removeHtmlElements(msg)

			// then
			Expect(sanitizedMsg).To(Equal("Custom message"))
		})

		It("should create default message with no-found-custom-file suffix using wrong url path", func() {
			// given
			gock.New("http://my.server.com").
				Get("path/to/custom_message_file.md").
				Reply(404)

			url := "http://github.com/my/repo/test-keeper.yaml"
			config := plugin.TestKeeperConfiguration{
				PluginConfiguration: config.PluginConfiguration{
					LocationURL: url,
					PluginHint:  "http://my.server.com/path/to/custom_message_file.md",
				},
			}

			// when
			msg := plugin.CreateCommentMessage(config, scm.RepositoryChange{})

			// then
			Expect(msg).NotTo(ContainSubstring("http://arquillian.org/ike-prow-plugins/#_test_keeper_plugin"))
			Expect(msg).To(ContainSubstring(url))
			Expect(msg).To(ContainSubstring(plugin.SkipComment))
			Expect(msg).To(ContainSubstring(
				"http://my.server.com/path/to/custom_message_file.md"))
		})

		It("should create default message with no-found-custom-file suffix using not-validate url", func() {
			// given
			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/path/to/custom_message_file.md").
				Reply(404)

			url := "http://github.com/my/repo/test-keeper.yaml"
			config := plugin.TestKeeperConfiguration{
				PluginConfiguration: config.PluginConfiguration{
					LocationURL: url,
					PluginHint:  "http/server.com/custom_message_file.md",
				},
			}

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			// when
			msg := plugin.CreateCommentMessage(config, change)

			// then
			Expect(msg).NotTo(ContainSubstring("http://arquillian.org/ike-prow-plugins/#_test_keeper_plugin"))
			Expect(msg).To(ContainSubstring(url))
			Expect(msg).To(ContainSubstring(plugin.SkipComment))
			Expect(msg).To(ContainSubstring(
				"https://raw.githubusercontent.com/owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/" +
					"http/server.com/custom_message_file.md"))
		})
	})
})

func removeHtmlElements(msg string) string {
	return strings.TrimSpace(string(bluemonday.StrictPolicy().SanitizeBytes([]byte(msg))))
}

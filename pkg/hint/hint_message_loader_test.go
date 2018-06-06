package hint_test

import (
	"github.com/arquillian/ike-prow-plugins/pkg/hint"
	"github.com/arquillian/ike-prow-plugins/pkg/config"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

const ProwPluginName string = "my-test-plugin"

var _ = Describe("Hint message creation", func() {

	Context("Creation of default hint messages that are sent to a validated PR when custom message file is not set", func() {

		It("should create default message referencing to documentation when url to config is empty", func() {
			// given
			conf := config.PluginConfiguration{PluginName: ProwPluginName}
			hint := &hint.Hint{
				Message: &hint.Message{ConfigFile: conf.LocationURL, Documentation: "#_my_test_plugin"},
			}

			// when
			msg := hint.LoadMessage(conf, scm.RepositoryChange{})

			// then
			Expect(msg).To(ContainSubstring("http://arquillian.org/ike-prow-plugins/#_my_test_plugin"))
			Expect(msg).NotTo(ContainSubstring("Your plugin configuration is stored in"))
		})

		It("should create default message referencing to config file when url to config is not empty", func() {
			// given
			url := "http://github.com/my/repo/my-test-plugin.yaml"
			conf := config.PluginConfiguration{LocationURL: url}
			hint := &hint.Hint{
				Message: &hint.Message{ConfigFile: conf.LocationURL, Documentation: "#_my_test_plugin"},
			}

			// when
			msg := hint.LoadMessage(conf, scm.RepositoryChange{})

			// then
			Expect(msg).NotTo(ContainSubstring("http://arquillian.org/ike-prow-plugins/#_my_test_plugin"))
			Expect(msg).To(ContainSubstring(url))
		})
	})

	Context("Creation of default hint messages from default location when config plugin hint is not set", func() {

		BeforeEach(func() {
			defer gock.OffAll()
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should create message taken from a default hint file if plugin hint url is missing", func() {
			// given
			NonExistingRawGitHubFiles("my-test-plugin.yaml", "my-test-plugin.yml")

			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/.ike-prow/my-test-plugin_hint.md").
				Reply(200).
				BodyString("Custom message")

			url := "http://github.com/my/repo/my-test-plugin.yaml"
			conf := config.PluginConfiguration{
				LocationURL: url,
				PluginName:  ProwPluginName,
			}

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			hint := &hint.Hint{
				Message: &hint.Message{ConfigFile: conf.LocationURL, Documentation: "#_my_test_plugin"},
			}

			// when
			msg := hint.LoadMessage(conf, change)

			// then
			Expect(msg).To(Equal("Custom message"))
		})

		It("should create message taken from a default hint file if location url is missing", func() {
			// given
			NonExistingRawGitHubFiles("my-test-plugin.yaml", "my-test-plugin.yml")

			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/.ike-prow/my-test-plugin_hint.md").
				Reply(200).
				BodyString("Custom message")

			conf := config.PluginConfiguration{PluginName: ProwPluginName}

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			hint := &hint.Hint{
				Message: &hint.Message{ConfigFile: conf.LocationURL, Documentation: "#_my_test_plugin"},
			}

			// when
			msg := hint.LoadMessage(conf, change)

			// then
			Expect(msg).To(Equal("Custom message"))
		})
	})

	Context("Creation of default hint messages that are sent to a validated PR when custom message file is set", func() {

		BeforeEach(func() {
			defer gock.OffAll()
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should create message taken from a file set in config using relative path", func() {
			// given
			NonExistingRawGitHubFiles("my-test-plugin.yaml", "my-test-plugin.yml")

			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/path/to/my-test-plugin_hint.md").
				Reply(200).
				BodyString("Custom message")

			url := "http://github.com/my/repo/my-test-plugin.yaml"

			conf := config.PluginConfiguration{
				LocationURL: url,
				PluginName:  ProwPluginName,
				PluginHint:  "path/to/my-test-plugin_hint.md",
			}

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			hint := &hint.Hint{
				Message: &hint.Message{ConfigFile: conf.LocationURL, Documentation: "#_my_test_plugin"},
			}

			// when
			msg := hint.LoadMessage(conf, change)

			// then
			Expect(msg).To(Equal("Custom message"))
		})

		It("should create default message with no-found-custom-file suffix using wrong relative path", func() {
			// given
			NonExistingRawGitHubFiles("path/to/my-test-plugin_hint.md", "my-test-plugin.yaml", "my-test-plugin.yml")

			url := "http://github.com/my/repo/my-test-plugin.yaml"
			conf := config.PluginConfiguration{
				LocationURL: url,
				PluginName:  ProwPluginName,
				PluginHint:  "path/to/my-test-plugin_hint.md",
			}

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			hint := &hint.Hint{
				Message: &hint.Message{ConfigFile: conf.LocationURL, Documentation: "#_my_test_plugin"},
			}

			// when
			msg := hint.LoadMessage(conf, change)

			// then
			Expect(msg).NotTo(ContainSubstring("http://arquillian.org/ike-prow-plugins/#_my_test_plugin"))
			Expect(msg).To(ContainSubstring(url))
			Expect(msg).To(ContainSubstring(
				"https://raw.githubusercontent.com/owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/" +
					"path/to/my-test-plugin_hint.md"))
		})

		It("should create message taken from a file set in config using url", func() {
			// given
			gock.New("http://my.server.com").
				Get("path/to/my-test-plugin_hint.md").Reply(200).
				BodyString("Custom message")

			url := "http://github.com/my/repo/my-test-plugin.yaml"
			conf := config.PluginConfiguration{
				LocationURL: url,
				PluginName:  ProwPluginName,
				PluginHint:  "http://my.server.com/path/to/my-test-plugin_hint.md",
			}

			hint := &hint.Hint{
				Message: &hint.Message{ConfigFile: conf.LocationURL, Documentation: "#_my_test_plugin"},
			}

			// when
			msg := hint.LoadMessage(conf, scm.RepositoryChange{})

			// then
			Expect(msg).To(Equal("Custom message"))
		})

		It("should create message taken from a file with upper case filename set in config using url", func() {
			// given
			gock.New("http://my.server.com").
				Get("path/to/MY-TEST-PLUGIN_HINT.MD").
				Reply(200).
				BodyString("Custom message")

			url := "http://github.com/my/repo/my-test-plugin.yaml"
			conf := config.PluginConfiguration{
				LocationURL: url,
				PluginName:  ProwPluginName,
				PluginHint:  "http://my.server.com/path/to/MY-TEST-PLUGIN_HINT.MD",
			}

			hint := &hint.Hint{
				Message: &hint.Message{ConfigFile: conf.LocationURL, Documentation: "#_my_test_plugin"},
			}

			// when
			msg := hint.LoadMessage(conf, scm.RepositoryChange{})

			// then
			Expect(msg).To(Equal("Custom message"))
		})

		It("should create default message with no-found-custom-file suffix using wrong url path", func() {
			// given
			gock.New("http://my.server.com").
				Get("path/to/my-test-plugin_hint.md").
				Reply(404)

			url := "http://github.com/my/repo/my-test-plugin.yaml"
			conf := config.PluginConfiguration{
				LocationURL: url,
				PluginName:  ProwPluginName,
				PluginHint:  "http://my.server.com/path/to/my-test-plugin_hint.md",
			}

			hint := &hint.Hint{
				Message: &hint.Message{ConfigFile: conf.LocationURL, Documentation: "#_my_test_plugin"},
			}

			// when
			msg := hint.LoadMessage(conf, scm.RepositoryChange{})

			// then
			Expect(msg).NotTo(ContainSubstring("http://arquillian.org/ike-prow-plugins/#_my_test_plugin"))
			Expect(msg).To(ContainSubstring(url))
			Expect(msg).To(ContainSubstring(
				"http://my.server.com/path/to/my-test-plugin_hint.md"))
		})

		It("should create default message with no-found-custom-file suffix using not-validate url", func() {
			// given
			NonExistingRawGitHubFiles("http/server.com/my-test-plugin_hint.md")

			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/path/to/my-test-plugin_hint.md").
				Reply(404)

			url := "http://github.com/my/repo/my-test-plugin.yaml"
			conf := config.PluginConfiguration{
				LocationURL: url,
				PluginName:  ProwPluginName,
				PluginHint:  "http/server.com/my-test-plugin_hint.md",
			}

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			hint := &hint.Hint{
				Message: &hint.Message{ConfigFile: conf.LocationURL, Documentation: "#_my_test_plugin"},
			}

			// when
			msg := hint.LoadMessage(conf, change)

			// then
			Expect(msg).NotTo(ContainSubstring("http://arquillian.org/ike-prow-plugins/#_my_test_plugin"))
			Expect(msg).To(ContainSubstring(url))
			Expect(msg).To(ContainSubstring(
				"https://raw.githubusercontent.com/owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/" +
					"http/server.com/my-test-plugin_hint.md"))
		})

		It("should create message taken from a string set in config", func() {
			// given
			url := "http://github.com/my/repo/my-test-plugin.yaml"
			conf := config.PluginConfiguration{
				LocationURL: url,
				PluginName:  ProwPluginName,
				PluginHint:  "Custom message",
			}

			hint := &hint.Hint{
				Message: &hint.Message{ConfigFile: conf.LocationURL, Documentation: "#_my_test_plugin"},
			}

			// when
			msg := hint.LoadMessage(conf, scm.RepositoryChange{})

			// then
			Expect(msg).To(Equal("Custom message"))
		})

		It("should create message with file path as content using invalid filename pattern", func() {
			// given
			gock.New("http://my.server.com").
				Get("path/to/custom_message_file.md").
				Reply(404)

			url := "http://github.com/my/repo/my-test-plugin.yaml"
			conf := config.PluginConfiguration{
				LocationURL: url,
				PluginName:  ProwPluginName,
				PluginHint:  "http://my.server.com/path/to/custom_message_file.md",
			}

			hint := &hint.Hint{
				Message: &hint.Message{ConfigFile: conf.LocationURL, Documentation: "#_my_test_plugin"},
			}

			// when
			msg := hint.LoadMessage(conf, scm.RepositoryChange{})

			// then
			Expect(msg).NotTo(ContainSubstring("http://arquillian.org/ike-prow-plugins/#_my_test_plugin"))
			Expect(msg).To(ContainSubstring("http://my.server.com/path/to/custom_message_file.md"))
		})
	})
})

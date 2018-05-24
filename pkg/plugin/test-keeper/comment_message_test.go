package testkeeper_test

import (
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/config"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper"
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
			conf := testkeeper.PluginConfiguration{PluginConfiguration: config.PluginConfiguration{PluginName: testkeeper.ProwPluginName}}

			// when
			msg := testkeeper.CreateCommentMessage(conf, scm.RepositoryChange{})

			// then
			Expect(msg).To(ContainSubstring("http://arquillian.org/ike-prow-plugins/#_test_keeper_plugin"))
			Expect(msg).To(ContainSubstring(testkeeper.BypassCheckComment))
		})

		It("should create default message referencing to config file when url to config is not empty", func() {
			// given
			url := "http://github.com/my/repo/test-keeper.yaml"
			conf := testkeeper.PluginConfiguration{PluginConfiguration: config.PluginConfiguration{LocationURL: url}}

			// when
			msg := testkeeper.CreateCommentMessage(conf, scm.RepositoryChange{})

			// then
			Expect(msg).NotTo(ContainSubstring("http://arquillian.org/ike-prow-plugins/#_test_keeper_plugin"))
			Expect(msg).To(ContainSubstring(url))
			Expect(msg).To(ContainSubstring(testkeeper.BypassCheckComment))
		})
	})

	Context("Creation of default comment messages from default location when config plugin hint is not set", func() {

		BeforeEach(func() {
			defer gock.OffAll()
		})

		It("should create message taken from a default hint file if plugin hint url is missing", func() {
			// given
			NonExistingRawGitHubFiles("test-keeper.yaml", "test-keeper.yml")

			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/.ike-prow/test-keeper_hint.md").
				Reply(200).
				BodyString("Custom message")

			url := "http://github.com/my/repo/test-keeper.yaml"
			conf := testkeeper.PluginConfiguration{
				PluginConfiguration: config.PluginConfiguration{
					LocationURL: url,
					PluginName:  testkeeper.ProwPluginName,
				},
			}

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			// when
			msg := testkeeper.CreateCommentMessage(conf, change)
			sanitizedMsg := removeHtmlElements(msg)

			// then
			Expect(sanitizedMsg).To(Equal("Custom message"))
		})

		It("should create message taken from a default hint file if location url is missing", func() {
			// given
			NonExistingRawGitHubFiles("test-keeper.yaml", "test-keeper.yml")

			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/.ike-prow/test-keeper_hint.md").
				Reply(200).
				BodyString("Custom message")

			conf := testkeeper.PluginConfiguration{
				PluginConfiguration: config.PluginConfiguration{
					PluginName:  testkeeper.ProwPluginName,
				},
			}

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			// when
			msg := testkeeper.CreateCommentMessage(conf, change)
			sanitizedMsg := removeHtmlElements(msg)

			// then
			Expect(sanitizedMsg).To(Equal("Custom message"))
		})
	})

	Context("Creation of default comment messages that are sent to a validated PR when custom message file is set", func() {

		BeforeEach(func() {
			defer gock.OffAll()
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should create message taken from a file set in config using relative path", func() {
			// given
			NonExistingRawGitHubFiles("test-keeper.yaml", "test-keeper.yml")

			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/path/to/test-keeper_hint.md").
				Reply(200).
				BodyString("Custom message")

			url := "http://github.com/my/repo/test-keeper.yaml"
			conf := testkeeper.PluginConfiguration{
				PluginConfiguration: config.PluginConfiguration{
					LocationURL: url,
					PluginName:  testkeeper.ProwPluginName,
					PluginHint:  "path/to/test-keeper_hint.md",
				},
			}

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			// when
			msg := testkeeper.CreateCommentMessage(conf, change)
			sanitizedMsg := removeHtmlElements(msg)

			// then
			Expect(sanitizedMsg).To(Equal("Custom message"))
		})

		It("should create default message with no-found-custom-file suffix using wrong relative path", func() {
			// given
			NonExistingRawGitHubFiles("path/to/test-keeper_hint.md", "test-keeper.yaml", "test-keeper.yml")

			url := "http://github.com/my/repo/test-keeper.yaml"
			conf := testkeeper.PluginConfiguration{
				PluginConfiguration: config.PluginConfiguration{
					LocationURL: url,
					PluginName:  testkeeper.ProwPluginName,
					PluginHint:  "path/to/test-keeper_hint.md",
				},
			}

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			// when
			msg := testkeeper.CreateCommentMessage(conf, change)

			// then
			Expect(msg).NotTo(ContainSubstring("http://arquillian.org/ike-prow-plugins/#_test_keeper_plugin"))
			Expect(msg).To(ContainSubstring(url))
			Expect(msg).To(ContainSubstring(testkeeper.BypassCheckComment))
			Expect(msg).To(ContainSubstring(
				"https://raw.githubusercontent.com/owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/" +
					"path/to/test-keeper_hint.md"))
		})

		It("should create message taken from a file set in config using url", func() {
			// given
			gock.New("http://my.server.com").
				Get("path/to/test-keeper_hint.md").Reply(200).
				BodyString("Custom message")

			url := "http://github.com/my/repo/test-keeper.yaml"
			conf := testkeeper.PluginConfiguration{
				PluginConfiguration: config.PluginConfiguration{
					LocationURL: url,
					PluginName:  testkeeper.ProwPluginName,
					PluginHint:  "http://my.server.com/path/to/test-keeper_hint.md",
				},
			}

			// when
			msg := testkeeper.CreateCommentMessage(conf, scm.RepositoryChange{})
			sanitizedMsg := removeHtmlElements(msg)

			// then
			Expect(sanitizedMsg).To(Equal("Custom message"))
		})

		It("should create message taken from a file with upper case filename set in config using url", func() {
			// given
			gock.New("http://my.server.com").
				Get("path/to/TEST-KEEPER_HINT.MD").
				Reply(200).
				BodyString("Custom message")

			url := "http://github.com/my/repo/test-keeper.yaml"
			conf := testkeeper.PluginConfiguration{
				PluginConfiguration: config.PluginConfiguration{
					LocationURL: url,
					PluginName:  testkeeper.ProwPluginName,
					PluginHint:  "http://my.server.com/path/to/TEST-KEEPER_HINT.MD",
				},
			}

			// when
			msg := testkeeper.CreateCommentMessage(conf, scm.RepositoryChange{})
			sanitizedMsg := removeHtmlElements(msg)

			// then
			Expect(sanitizedMsg).To(Equal("Custom message"))
		})

		It("should create default message with no-found-custom-file suffix using wrong url path", func() {
			// given
			gock.New("http://my.server.com").
				Get("path/to/test-keeper_hint.md").
				Reply(404)

			url := "http://github.com/my/repo/test-keeper.yaml"
			conf := testkeeper.PluginConfiguration{
				PluginConfiguration: config.PluginConfiguration{
					LocationURL: url,
					PluginName:  testkeeper.ProwPluginName,
					PluginHint:  "http://my.server.com/path/to/test-keeper_hint.md",
				},
			}

			// when
			msg := testkeeper.CreateCommentMessage(conf, scm.RepositoryChange{})

			// then
			Expect(msg).NotTo(ContainSubstring("http://arquillian.org/ike-prow-plugins/#_test_keeper_plugin"))
			Expect(msg).To(ContainSubstring(url))
			Expect(msg).To(ContainSubstring(testkeeper.BypassCheckComment))
			Expect(msg).To(ContainSubstring(
				"http://my.server.com/path/to/test-keeper_hint.md"))
		})

		It("should create default message with no-found-custom-file suffix using not-validate url", func() {
			// given
			NonExistingRawGitHubFiles("http/server.com/test-keeper_hint.md")

			gock.New("https://raw.githubusercontent.com").
				Get("owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/path/to/test-keeper_hint.md").
				Reply(404)

			url := "http://github.com/my/repo/test-keeper.yaml"
			conf := testkeeper.PluginConfiguration{
				PluginConfiguration: config.PluginConfiguration{
					LocationURL: url,
					PluginName:  testkeeper.ProwPluginName,
					PluginHint:  "http/server.com/test-keeper_hint.md",
				},
			}

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}

			// when
			msg := testkeeper.CreateCommentMessage(conf, change)

			// then
			Expect(msg).NotTo(ContainSubstring("http://arquillian.org/ike-prow-plugins/#_test_keeper_plugin"))
			Expect(msg).To(ContainSubstring(url))
			Expect(msg).To(ContainSubstring(testkeeper.BypassCheckComment))
			Expect(msg).To(ContainSubstring(
				"https://raw.githubusercontent.com/owner/repo/46cb8fac44709e4ccaae97448c65e8f7320cfea7/" +
					"http/server.com/test-keeper_hint.md"))
		})

		It("should create message taken from a string set in config", func() {
			// given
			url := "http://github.com/my/repo/test-keeper.yaml"
			conf := testkeeper.PluginConfiguration{
				PluginConfiguration: config.PluginConfiguration{
					LocationURL: url,
					PluginName:  testkeeper.ProwPluginName,
					PluginHint:  "Custom message",
				},
			}

			// when
			msg := testkeeper.CreateCommentMessage(conf, scm.RepositoryChange{})

			// then
			Expect(msg).To(Equal("Custom message"))
		})

		It("should create message with file path as content using invalid filename pattern", func() {
			// given
			gock.New("http://my.server.com").
				Get("path/to/custom_message_file.md").
				Reply(404)

			url := "http://github.com/my/repo/test-keeper.yaml"
			conf := testkeeper.PluginConfiguration{
				PluginConfiguration: config.PluginConfiguration{
					LocationURL: url,
					PluginName:  testkeeper.ProwPluginName,
					PluginHint:  "http://my.server.com/path/to/custom_message_file.md",
				},
			}

			// when
			msg := testkeeper.CreateCommentMessage(conf, scm.RepositoryChange{})

			// then
			Expect(msg).NotTo(ContainSubstring("http://arquillian.org/ike-prow-plugins/#_test_keeper_plugin"))
			Expect(msg).To(ContainSubstring("http://my.server.com/path/to/custom_message_file.md"))
		})
	})
})

func removeHtmlElements(msg string) string {
	return strings.TrimSpace(string(bluemonday.StrictPolicy().SanitizeBytes([]byte(msg))))
}

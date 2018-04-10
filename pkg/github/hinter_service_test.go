package github_test

import (
	"github.com/arquillian/ike-prow-plugins/pkg/github"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("Config loader features", func() {

	Context("Loading configuration file from the repository", func() {

		var client *github.Client

		BeforeEach(func() {
			gock.Off()

			client = NewDefaultGitHubClient()
		})

		It("should add new comment with main title, dev mention and plugin message when no such a comment exists", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/issues/2/comments").
				Reply(200).
				BodyString("{}")

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}
			hintContext := github.HintContext{
				PluginName: "my-plugin-name",
				Assignee:   "toAssign",
			}
			hinter := github.NewHinter(client, NewDiscardOutLogger(), change, 2, hintContext)

			toHaveBodyWithWholePluginsComment := func(statusPayload map[string]interface{}) bool {
				return Expect(statusPayload).To(SatisfyAll(
					HaveBodyThatContains("### Ike Plugins (my-plugin-name)"),
					HaveBodyThatContains("@toAssign"),
					HaveBodyThatContains("New comment"),
				))
			}

			gock.New("https://api.github.com").
				Post("/repos/owner/repo/issues/2/comments").
				SetMatcher(ExpectPayload(toHaveBodyWithWholePluginsComment)).
				Reply(201)

			// when
			err := hinter.PluginComment("New comment")

			// then
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should not send any request when message from the plugin exists", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/issues/2/comments").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/comments_with_tests_keeper_message.json"))

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}
			hintContext := github.HintContext{
				PluginName: "test-keeper",
				Assignee:   "toAssign",
			}

			hinter := github.NewHinter(client, NewDiscardOutLogger(), change, 2, hintContext)

			// when
			err := hinter.PluginComment("New comment")

			// then
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should create a new comment that contains missing plugin hint when different one already exists", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/issues/2/comments").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/comments_with_tests_keeper_message.json"))

			change := scm.RepositoryChange{
				Owner:    "owner",
				RepoName: "repo",
				Hash:     "46cb8fac44709e4ccaae97448c65e8f7320cfea7",
			}
			hintContext := github.HintContext{
				PluginName: "another-plugin",
				Assignee:   "toAssign",
			}

			expContent := "### Ike Plugins (another-plugin)\n\nThank you @toAssign for this contribution!" +
				"\n\nNew comment"

			toHaveModifiedBody := func(statusPayload map[string]interface{}) bool {
				return Expect(statusPayload).To(SatisfyAll(
					HaveBody(expContent),
				))
			}

			gock.New("https://api.github.com").
				Post("/repos/owner/repo/issues/2/comments").
				SetMatcher(ExpectPayload(toHaveModifiedBody)).
				Reply(200)

			hinter := github.NewHinter(client, NewDiscardOutLogger(), change, 2, hintContext)

			// when
			err := hinter.PluginComment("New comment")

			// then
			Ω(err).ShouldNot(HaveOccurred())
		})
	})
})

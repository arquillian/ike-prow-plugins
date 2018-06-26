package message_test

import (
	"github.com/arquillian/ike-prow-plugins/pkg/config"
	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/status/message"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("Config loader features", func() {

	Context("Loading configuration file from the repository", func() {

		var client ghclient.Client

		BeforeEach(func() {
			defer gock.OffAll()

			client = NewDefaultGitHubClient()
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should add new comment with main title, dev mention and plugin message when no such a comment exists", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/issues/2/comments").
				Reply(200).
				BodyString("[]")

			commentsLoader := &ghservice.IssueCommentsLazyLoader{
				Client: client,
				Issue:  *scm.NewRepositoryIssue("owner", "repo", 2),
			}
			messageContext := message.NewStatusMessageContext("my-plugin-name", "docSection",
				NewPullRequest("owner", "repo", "1a2b", "toAssign"), config.PluginConfiguration{})
			msgService := message.NewStatusMessageService(client, log.NewTestLogger(), commentsLoader, messageContext)

			toHaveBodyWithWholePluginsComment := SoftlySatisfyAll(
				HaveBodyThatContains("### Ike Plugins (my-plugin-name)"),
				HaveBodyThatContains("@toAssign"),
				HaveBodyThatContains("New comment"),
			)

			gock.New("https://api.github.com").
				Post("/repos/owner/repo/issues/2/comments").
				SetMatcher(ExpectPayload(toHaveBodyWithWholePluginsComment)).
				Reply(201)

			// when
			err := msgService.StatusMessage(newBasicMsgCreator("New comment"), true)

			// then
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should not send any request when message from the plugin exists", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/issues/2/comments").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/comments_with_tests_keeper_message.json"))

			commentsLoader := &ghservice.IssueCommentsLazyLoader{
				Client: client,
				Issue:  *scm.NewRepositoryIssue("owner", "repo", 2),
			}
			messageContext := message.NewStatusMessageContext("test-keeper", "docSection",
				NewPullRequest("owner", "repo", "1a2b", "toAssign"), config.PluginConfiguration{})

			msgService := message.NewStatusMessageService(client, log.NewTestLogger(), commentsLoader, messageContext)

			// when
			err := msgService.StatusMessage(newBasicMsgCreator("Message from test-keeper"), true)

			// then
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should create a new comment that contains missing status message when different one already exists", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/issues/2/comments").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/comments_with_tests_keeper_message.json"))

			commentsLoader := &ghservice.IssueCommentsLazyLoader{
				Client: client,
				Issue:  *scm.NewRepositoryIssue("owner", "repo", 2),
			}
			messageContext := message.NewStatusMessageContext("another-plugin", "docSection",
				NewPullRequest("owner", "repo", "1a2b", "toAssign"), config.PluginConfiguration{})

			expContent := "### Ike Plugins (another-plugin)\n\nThank you @toAssign for this contribution!" +
				"\n\nNew comment"

			toHaveModifiedBody := SoftlySatisfyAll(
				HaveBody(expContent),
			)

			gock.New("https://api.github.com").
				Post("/repos/owner/repo/issues/2/comments").
				SetMatcher(ExpectPayload(toHaveModifiedBody)).
				Reply(200)

			msgService := message.NewStatusMessageService(client, log.NewTestLogger(), commentsLoader, messageContext)

			// when
			err := msgService.StatusMessage(newBasicMsgCreator("New comment"), true)

			// then
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should modify existing comment when a different content of the status message is provided", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/issues/2/comments").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/comments_with_tests_keeper_message.json"))

			commentsLoader := &ghservice.IssueCommentsLazyLoader{
				Client: client,
				Issue:  *scm.NewRepositoryIssue("owner", "repo", 2),
			}
			messageContext := message.NewStatusMessageContext("test-keeper", "docSection",
				NewPullRequest("owner", "repo", "1a2b", "toAssign"), config.PluginConfiguration{})

			toHaveBodyWithWholePluginsComment := SoftlySatisfyAll(
				HaveBodyThatContains("### Ike Plugins (test-keeper)"),
				HaveBodyThatContains("@toAssign"),
				HaveBodyThatContains("New Message"),
			)

			gock.New("https://api.github.com").
				Patch("/repos/owner/repo/issues/comments/372707978").
				SetMatcher(ExpectPayload(toHaveBodyWithWholePluginsComment)).
				Reply(201)

			msgService := message.NewStatusMessageService(client, log.NewTestLogger(), commentsLoader, messageContext)

			// when
			err := msgService.StatusMessage(newBasicMsgCreator("New Message"), true)

			// then
			Ω(err).ShouldNot(HaveOccurred())
		})
	})
})

func newBasicMsgCreator(msg string) func() string {
	return func() string {
		return msg
	}
}

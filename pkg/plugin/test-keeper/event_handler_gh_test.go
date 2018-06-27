package testkeeper_test

import (
	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/command"
	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

const botName = "alien-ike"

var _ = Describe("Test Keeper Plugin features", func() {

	var handler *testkeeper.GitHubTestEventsHandler
	var mocker = NewMockPluginTemplate(testkeeper.ProwPluginName)

	log := log.NewTestLogger()

	Context("Pull Request event handling", func() {

		BeforeEach(func() {
			defer gock.OffAll()
			handler = &testkeeper.GitHubTestEventsHandler{Client: NewDefaultGitHubClient(), BotName: botName}
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should approve opened pull request and update status message when tests included", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithFiles(LoadedFrom("test_fixtures/github_calls/prs/with_tests/changes.json")).
				WithComments(LoadedFrom("test_fixtures/github_calls/prs/comments_with_no_test_status_msg.json")).
				WithoutConfigFiles().
				WithoutMessageFiles("test-keeper_with_tests_message.md").
				Expecting(
					Status(ToBe(github.StatusSuccess, testkeeper.TestsExistMessage, testkeeper.TestsExistDetailsPageName)),
					ChangedComment(397622617, ContainingStatusMessage(testkeeper.WithTestsMsg))).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("opened"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should approve opened pull request when tests included based on configured pattern and defaults (implicitly combined)", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithFiles(LoadedFrom("test_fixtures/github_calls/prs/with_tests/changes_go_files.json")).
				WithoutComments().
				WithConfigFile(
					ConfigYml(Containing(
						Param("test_patterns", "['**/*_test_suite.go']"),
						Param("skip_validation_for", "['README.adoc']")))).
				Expecting(
					Status(ToBe(github.StatusSuccess, testkeeper.TestsExistMessage, testkeeper.TestsExistDetailsPageName))).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("opened"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should approve new pull request without tests when it comes with configuration excluding all files from test presence check (implicitly combined)", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithFiles(LoadedFrom("test_fixtures/github_calls/prs/without_tests/changes-with-test-keeper-config-excluding-other-file-from-PR.json")).
				WithConfigFile(
					ConfigYml(Containing(
						Param("skip_validation_for", "'**/Randomfile'")))).
				WithoutComments().
				Expecting(
					Status(ToBe(github.StatusSuccess, testkeeper.OkOnlySkippedFilesMessage, testkeeper.OkOnlySkippedFilesDetailsPageName)),
					Comment(ContainingStatusMessage(testkeeper.WithoutTestsMsg))).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("opened"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should reject opened pull request when no tests are matching defined pattern with no defaults implicitly combined", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithFiles(LoadedFrom("test_fixtures/github_calls/prs/with_tests/changes_go_files.json")).
				WithoutRawFiles(ghservice.ConfigHome+"test-keeper_hint.md").
				WithConfigFile(
					ConfigYml(LoadedFrom("test_fixtures/github_calls/prs/with_tests/test-keeper.yml"))).
				WithoutMessageFiles("test-keeper_without_tests_message.md").
				WithoutComments().
				Expecting(
					Status(ToBe(github.StatusFailure, testkeeper.NoTestsMessage, testkeeper.NoTestsDetailsPageName)),
					Comment(ContainingStatusMessage(testkeeper.WithoutTestsMsg))).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("opened"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should block newly created pull request when no tests are included", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithFiles(LoadedFrom("test_fixtures/github_calls/prs/without_tests/changes.json")).
				WithoutConfigFiles().
				WithoutMessageFiles("test-keeper_without_tests_message.md").
				WithoutComments().
				Expecting(
					Status(ToBe(github.StatusFailure, testkeeper.NoTestsMessage, testkeeper.NoTestsDetailsPageName)),
					Comment(ContainingStatusMessage(testkeeper.WithoutTestsMsg))).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("opened"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should not block newly created pull request when documentation and build files are the only changes", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithFiles(LoadedFrom("test_fixtures/github_calls/prs/without_tests/build_and_docs_only_changes.json")).
				WithoutConfigFiles().
				WithoutComments().
				Expecting(
					Status(ToBe(github.StatusSuccess, testkeeper.OkOnlySkippedFilesMessage, testkeeper.OkOnlySkippedFilesDetailsPageName))).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("opened"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should block newly created pull request when deletions in the tests are the only changes", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithFiles(LoadedFrom("test_fixtures/github_calls/prs/without_tests/deletions_only_changes_in_tests.json")).
				WithoutComments().
				WithoutConfigFiles().
				WithoutMessageFiles("test-keeper_without_tests_message.md").
				Expecting(
					Comment(ContainingStatusMessage(testkeeper.WithoutTestsMsg)),
					Status(ToBe(github.StatusFailure, testkeeper.NoTestsMessage, testkeeper.NoTestsDetailsPageName))).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("opened"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should block newly created pull request when there are changes in the business logic but only deletions in the tests", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithFiles(LoadedFrom("test_fixtures/github_calls/prs/without_tests/prod_code_changes_with_deletion_only_in_tests.json")).
				WithoutComments().
				WithoutConfigFiles().
				WithoutMessageFiles("test-keeper_without_tests_message.md").
				Expecting(
					Comment(ContainingStatusMessage(testkeeper.WithoutTestsMsg)),
					Status(ToBe(github.StatusFailure, testkeeper.NoTestsMessage, testkeeper.NoTestsDetailsPageName))).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("opened"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should send ok status when PR contains no test but a comment with bypass command is present", func() {
			approvedBy := fmt.Sprintf(testkeeper.ApprovedByMessage, "bartoszmajsak")

			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithFiles(LoadedFrom("test_fixtures/github_calls/prs/without_tests/changes.json")).
				WithUsers(Admin("bartoszmajsak")).
				WithComments(`[{"user":{"login":"bartoszmajsak"}, "body":"` + testkeeper.BypassCheckComment + `"}]`).
				WithoutReviews().
				WithoutConfigFiles().
				Expecting(
					Status(ToBe(github.StatusSuccess, approvedBy, testkeeper.ApprovedByDetailsPageName))).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("opened"))

			// then - should not expect any additional request mocking
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should block pull request without tests and with comments containing bypass message added by user with insufficient permissions", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithFiles(LoadedFrom("test_fixtures/github_calls/prs/without_tests/changes.json")).
				WithUsers(ExternalUser("bartoszmajsak-test")).
				WithComments(LoadedFrom("test_fixtures/github_calls/prs/comments_with_no_test_status_msg.json")).
				WithoutReviews().
				WithoutConfigFiles().
				WithoutMessageFiles("test-keeper_without_tests_message.md").
				Expecting(
					Status(ToBe(github.StatusFailure, testkeeper.NoTestsMessage, testkeeper.NoTestsDetailsPageName))).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("opened"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Context("Pull Request comment event handling", func() {

		BeforeEach(func() {
			defer gock.OffAll()
			handler = &testkeeper.GitHubTestEventsHandler{Client: NewDefaultGitHubClient(), BotName: botName}
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should skip test existence check when "+testkeeper.BypassCheckComment+" command is used by admin user", func() {
			// given
			approvedBy := fmt.Sprintf(testkeeper.ApprovedByMessage, "bartoszmajsak")
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithUsers(Admin("bartoszmajsak")).
				WithoutReviews().
				WithoutConfigFiles().
				Expecting(
					Status(ToBe(github.StatusSuccess, approvedBy, testkeeper.ApprovedByDetailsPageName))).
				Create()

			event := prMock.CreateCommentEvent(SentBy("bartoszmajsak"), testkeeper.BypassCheckComment, "created")

			// when
			err := handler.HandleIssueCommentEvent(log, event)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should ignore "+testkeeper.BypassCheckComment+" when used by non-admin user", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithUsers(ExternalUser("bartoszmajsak-test")).
				WithoutReviews().
				WithoutConfigFiles().
				Expecting(
					Comment(To(
						HaveBodyThatContains("Hey @bartoszmajsak-test! It seems you tried to trigger `/ok-without-tests` command"),
						HaveBodyThatContains("You have to be admin or requested reviewer or pull request approver, but not pull request creator"))),
					Status(ToBe(github.StatusFailure, testkeeper.NoTestsMessage, testkeeper.NoTestsDetailsPageName))).
				Create()

			event := prMock.CreateCommentEvent(SentBy("bartoszmajsak-test"), testkeeper.BypassCheckComment, "created")

			// when
			err := handler.HandleIssueCommentEvent(log, event)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Context("Trigger test-keeper plugin by triggering comment on pull request", func() {
		BeforeEach(func() {
			defer gock.OffAll()
			handler = &testkeeper.GitHubTestEventsHandler{Client: NewDefaultGitHubClient(), BotName: botName}
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should block newly created pull request without tests when "+command.RunCommentPrefix+" all command is used by admin user", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithFiles(LoadedFrom("test_fixtures/github_calls/prs/without_tests/changes.json")).
				WithoutComments().
				WithUsers(Admin("bartoszmajsak")).
				WithoutReviews().
				WithoutConfigFiles().
				WithoutMessageFiles("test-keeper_without_tests_message.md").
				Expecting(Comment(ContainingStatusMessage(testkeeper.WithoutTestsMsg)),
					Status(ToBe(github.StatusFailure, testkeeper.NoTestsMessage, testkeeper.NoTestsDetailsPageName))).
				Create()

			event := prMock.CreateCommentEvent(SentBy("bartoszmajsak"), "/run all", "created")

			// when
			err := handler.HandleIssueCommentEvent(log, event)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should approve newly created pull request with tests when "+command.RunCommentPrefix+" "+testkeeper.ProwPluginName+" command is triggered by pr reviewer", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithFiles(LoadedFrom("test_fixtures/github_calls/prs/with_tests/changes.json")).
				WithoutComments().
				WithUsers(ExternalUser("bartoszmajsak-test"), RequestedReviewer("bartoszmajsak-test")).
				WithoutReviews().
				WithoutConfigFiles().
				Expecting(Comment(ContainingStatusMessage(testkeeper.WithoutTestsMsg)),
					Status(ToBe(github.StatusSuccess, testkeeper.TestsExistMessage, testkeeper.TestsExistDetailsPageName))).
				Create()

			event := prMock.CreateCommentEvent(SentByReviewer, "/run test-keeper", "created")

			// when
			err := handler.HandleIssueCommentEvent(log, event)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should do nothing for newly created pull request with tests when "+command.RunCommentPrefix+" work-in-progress command is triggered by pr reviewer", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithFiles(LoadedFrom("test_fixtures/github_calls/prs/with_tests/changes.json")).
				WithoutComments().
				WithUsers(ExternalUser("bartoszmajsak-test"), RequestedReviewer("bartoszmajsak-test")).
				WithoutReviews().
				WithoutConfigFiles().
				Expecting(
					NoStatus(),
					NoComment()).
				Create()

			event := prMock.CreateCommentEvent(SentByReviewer, "/run work-in-progress", "created")

			// when
			err := handler.HandleIssueCommentEvent(log, event)

			// then
			Ω(err).ShouldNot(HaveOccurred())
		})
	})
})

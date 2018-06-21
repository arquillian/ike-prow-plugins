package testkeeper_test

import (
	"fmt"
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/command"
	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper"
	gogh "github.com/google/go-github/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

const (
	repositoryName = "bartoszmajsak/wfswarm-booster-pipeline-test"
	botName        = "alien-ike"
)

var (
	expectedContext = strings.Join([]string{botName, testkeeper.ProwPluginName}, "/")
	docStatusRoot   = fmt.Sprintf("%s/status/%s", plugin.DocumentationURL, testkeeper.ProwPluginName)
)

var _ = Describe("Test Keeper Plugin features", func() {

	var handler *testkeeper.GitHubTestEventsHandler

	log := log.NewTestLogger()
	configFilePath := ghservice.ConfigHome + testkeeper.ProwPluginName

	toBe := func(status, description, context, detailsLink string) SoftMatcher {
		return SoftlySatisfyAll(
			HaveState(status),
			HaveDescription(description),
			HaveContext(context),
			HaveTargetURL(fmt.Sprintf("%s/%s/%s.html", docStatusRoot, strings.ToLower(status), detailsLink)),
		)

	}

	toHaveBodyWithWholePluginsComment := SoftlySatisfyAll(
		HaveBodyThatContains(fmt.Sprintf(ghservice.PluginTitleTemplate, testkeeper.ProwPluginName)),
		HaveBodyThatContains("@bartoszmajsak"),
	)

	Context("Pull Request event handling", func() {

		BeforeEach(func() {
			defer gock.OffAll()
			handler = &testkeeper.GitHubTestEventsHandler{Client: NewDefaultGitHubClient(), BotName: botName}
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should approve opened pull request when tests included", func() {
			// given
			NonExistingRawGitHubFiles("test-keeper.yml", "test-keeper.yaml", "test-keeper_hint.md")
			gockEmptyComments(2)

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/2/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/with_tests/changes.json"))

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toBe(github.StatusSuccess, testkeeper.TestsExistMessage, expectedContext, testkeeper.TestsExistDetailsPageName))).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/with_tests/status_opened.json")
			pullRequestEvent := TriggerPullRequestEvent(statusPayload, gogh.PullRequestEvent{})

			// when
			err := handler.HandlePullRequestEvent(log, pullRequestEvent)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should approve opened pull request when tests included based on configured pattern and defaults (implicitly combined)", func() {
			// given
			gockEmptyComments(2)

			gock.New("https://raw.githubusercontent.com").
				Get(repositoryName + "/5d6e9b25da90edfc19f488e595e0645c081c1575/" + configFilePath + ".yml").
				Reply(200).
				BodyString("test_patterns: ['**/*_test_suite.go']\n" +
					"skip_validation_for: ['README.adoc']")

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/2/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/with_tests/changes_go_files.json"))

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toBe(github.StatusSuccess, testkeeper.TestsExistMessage, expectedContext, testkeeper.TestsExistDetailsPageName))).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/with_tests/status_opened.json")
			pullRequestEvent := TriggerPullRequestEvent(statusPayload, gogh.PullRequestEvent{})

			// when
			err := handler.HandlePullRequestEvent(log, pullRequestEvent)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should approve new pull request without tests when it comes with configuration excluding all files from test presence check (implicitly combined)", func() {
			// given
			gockEmptyComments(1)

			gock.New("https://raw.githubusercontent.com").
				Get(repositoryName + "/5d6e9b25da90edfc19f488e595e0645c081c1575/" + configFilePath + ".yml").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/test-keeper-ignore-randomfile.yml"))

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/1/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/changes-with-test-keeper-config-excluding-other-file-from-PR.json"))

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toBe(github.StatusSuccess, testkeeper.OkOnlySkippedFilesMessage, expectedContext, testkeeper.OkOnlySkippedFilesDetailsPageName))).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/without_tests/status_opened.json")
			pullRequestEvent := TriggerPullRequestEvent(statusPayload, gogh.PullRequestEvent{})

			// when
			err := handler.HandlePullRequestEvent(log, pullRequestEvent)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should reject opened pull request when no tests are matching defined pattern with no defaults implicitly combined", func() {
			// given

			NonExistingRawGitHubFiles("test-keeper_hint.md")
			gockEmptyComments(2)

			gock.New("https://raw.githubusercontent.com").
				Get(repositoryName + "/5d6e9b25da90edfc19f488e595e0645c081c1575/" + configFilePath + ".yml").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/with_tests/test-keeper.yml"))

			gock.New("https://api.github.com").
				Get("/repos/"+repositoryName+"/pulls/2/files").
				MatchParam("per_page", "100").
				MatchParam("page", "1").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/with_tests/changes_go_files.json"))

			// This way we implicitly verify that call happened after `HandleEvent` call
			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toBe(github.StatusFailure, testkeeper.NoTestsMessage, expectedContext, testkeeper.NoTestsDetailsPageName))).
				Reply(201)
			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/issues/2/comments").
				SetMatcher(ExpectPayload(toHaveBodyWithWholePluginsComment)).
				Reply(201)

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/with_tests/status_opened.json")
			pullRequestEvent := TriggerPullRequestEvent(statusPayload, gogh.PullRequestEvent{})

			// when
			err := handler.HandlePullRequestEvent(log, pullRequestEvent)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should block newly created pull request when no tests are included", func() {
			// given
			NonExistingRawGitHubFiles("test-keeper.yml", "test-keeper.yaml", "test-keeper_hint.md")
			gockEmptyComments(1)

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/1/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/changes.json"))

			// This way we implicitly verify that call happened after `HandleEvent` call
			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toBe(github.StatusFailure, testkeeper.NoTestsMessage, expectedContext, testkeeper.NoTestsDetailsPageName))).
				Reply(201)
			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/issues/1/comments").
				SetMatcher(ExpectPayload(toHaveBodyWithWholePluginsComment)).
				Reply(201)

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/without_tests/status_opened.json")
			pullRequestEvent := TriggerPullRequestEvent(statusPayload, gogh.PullRequestEvent{})

			// when
			err := handler.HandlePullRequestEvent(log, pullRequestEvent)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should not block newly created pull request when documentation and build files are the only changes", func() {
			// given
			NonExistingRawGitHubFiles("test-keeper.yml", "test-keeper.yaml")
			gockEmptyComments(1)

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/1/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/build_and_docs_only_changes.json"))

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toBe(github.StatusSuccess, testkeeper.OkOnlySkippedFilesMessage, expectedContext, testkeeper.OkOnlySkippedFilesDetailsPageName))).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/without_tests/status_opened.json")
			pullRequestEvent := TriggerPullRequestEvent(statusPayload, gogh.PullRequestEvent{})

			// when
			err := handler.HandlePullRequestEvent(log, pullRequestEvent)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should block newly created pull request when deletions in the tests are the only changes", func() {
			// given
			NonExistingRawGitHubFiles("test-keeper.yml", "test-keeper.yaml", "test-keeper_hint.md")
			gockEmptyComments(1)

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/1/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/deletions_only_changes_in_tests.json"))

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toBe(github.StatusFailure, testkeeper.NoTestsMessage, expectedContext, testkeeper.NoTestsDetailsPageName))).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/issues/1/comments").
				SetMatcher(ExpectPayload(toHaveBodyWithWholePluginsComment)).
				Reply(201)

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/without_tests/status_opened.json")
			pullRequestEvent := TriggerPullRequestEvent(statusPayload, gogh.PullRequestEvent{})

			// when
			err := handler.HandlePullRequestEvent(log, pullRequestEvent)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should block newly created pull request when there are changes in the business logic but only deletions in the tests", func() {
			// given
			NonExistingRawGitHubFiles("test-keeper.yml", "test-keeper.yaml", "test-keeper_hint.md")
			gockEmptyComments(1)

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/1/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/prod_code_changes_with_deletion_only_in_tests.json"))

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toBe(github.StatusFailure, testkeeper.NoTestsMessage, expectedContext, testkeeper.NoTestsDetailsPageName))).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/issues/1/comments").
				SetMatcher(ExpectPayload(toHaveBodyWithWholePluginsComment)).
				Reply(201)

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/without_tests/status_opened.json")
			pullRequestEvent := TriggerPullRequestEvent(statusPayload, gogh.PullRequestEvent{})

			// when
			err := handler.HandlePullRequestEvent(log, pullRequestEvent)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should send ok status when PR contains no test but a comment with bypass command is present", func() {
			//given
			NonExistingRawGitHubFiles("test-keeper.yml", "test-keeper.yaml")
			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/issues/1/comments").
				Reply(200).
				BodyString(`[{"user":{"login":"bartoszmajsak"}, "body":"` + testkeeper.BypassCheckComment + `"}]`)

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/1/reviews").
				Reply(200).
				BodyString(`[]`)

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/1/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/changes.json"))

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/collaborators/bartoszmajsak/permission").
				Reply(200).
				BodyString(`{"permission": "admin"}`)

			toHaveEnforcedSuccessState := SoftlySatisfyAll(
				HaveState(github.StatusSuccess),
				HaveDescription(fmt.Sprintf(testkeeper.ApprovedByMessage, "bartoszmajsak")),
			)

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toHaveEnforcedSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/without_tests/status_opened_by_external_user.json")
			pullRequestEvent := TriggerPullRequestEvent(statusPayload, gogh.PullRequestEvent{})

			// when
			err := handler.HandlePullRequestEvent(log, pullRequestEvent)

			// then - should not expect any additional request mocking
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should block pull request without tests and with comments containing bypass message added by user with insufficient permissions", func() {
			// given
			NonExistingRawGitHubFiles("test-keeper.yml", "test-keeper.yaml", "test-keeper_hint.md")

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/issues/1/comments").
				Reply(200).
				BodyString(`[{"user":{"login":"bartoszmajsak-test"}, "body":"` + testkeeper.BypassCheckComment + `"},` +
					`{"body":"` + fmt.Sprintf(ghservice.PluginTitleTemplate, testkeeper.ProwPluginName) + `"}]`)

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/1/reviews").
				Reply(200).
				BodyString(`[]`)

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/collaborators/bartoszmajsak-test/permission").
				Reply(200).
				BodyString(`{"permission": "read"}`)

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/1/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/changes.json"))

			// This way we implicitly verify that call happened after `HandleEvent` call
			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toBe(github.StatusFailure, testkeeper.NoTestsMessage, expectedContext, testkeeper.NoTestsDetailsPageName))).
				Reply(201)

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/without_tests/status_opened_by_external_user.json")
			pullRequestEvent := TriggerPullRequestEvent(statusPayload, gogh.PullRequestEvent{})

			// when
			err := handler.HandlePullRequestEvent(log, pullRequestEvent)

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
			NonExistingRawGitHubFiles("test-keeper.yml", "test-keeper.yaml")

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/1").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/pr_details.json"))

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/collaborators/bartoszmajsak/permission").
				Reply(200).
				BodyString(`{"permission": "admin"}`)

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/1/reviews").
				Reply(200).
				BodyString(`[]`)

			toHaveEnforcedSuccessState := SoftlySatisfyAll(
				HaveState(github.StatusSuccess),
				HaveDescription(fmt.Sprintf(testkeeper.ApprovedByMessage, "bartoszmajsak")),
			)

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toHaveEnforcedSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/without_tests/skip_comment_by_admin.json")
			issueCommentEvent := TriggerIssueCommentEvent(statusPayload, gogh.IssueCommentEvent{})

			// when
			err := handler.HandleIssueCommentEvent(log, issueCommentEvent)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should ignore "+testkeeper.BypassCheckComment+" when used by non-admin user", func() {
			// given
			NonExistingRawGitHubFiles("test-keeper.yml", "test-keeper.yaml")

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/1").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/pr_details.json"))

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/collaborators/bartoszmajsak-test/permission").
				Reply(200).
				BodyString(`{"permission": "read"}`)

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toBe(github.StatusFailure, testkeeper.NoTestsMessage, expectedContext, testkeeper.NoTestsDetailsPageName))).
				Reply(201)

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/1/reviews").
				Reply(200).
				BodyString(`[]`)

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/issues/1/comments").
				SetMatcher(
					ExpectPayload(To(
							HaveBodyThatContains("Hey @bartoszmajsak-test! It seems you tried to trigger `/ok-without-tests` command"),
							HaveBodyThatContains("You have to be admin or requested reviewer or pull request approver, but not pull request creator")))).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/without_tests/skip_comment_by_external.json")
			issueCommentEvent := TriggerIssueCommentEvent(statusPayload, gogh.IssueCommentEvent{})

			// when
			err := handler.HandleIssueCommentEvent(log, issueCommentEvent)

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
			NonExistingRawGitHubFiles("test-keeper.yml", "test-keeper.yaml", "test-keeper_hint.md")

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/1").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/pr_details.json"))

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/1/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/changes.json"))

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/issues/1/comments").
				Reply(200).
				BodyString("[]")

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/collaborators/bartoszmajsak/permission").
				Reply(200).
				BodyString(`{"permission": "admin"}`)

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/1/reviews").
				Reply(200).
				BodyString(`[]`)

			// This way we implicitly verify that call happened after `HandleEvent` call
			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/issues/1/comments").
				SetMatcher(ExpectPayload(toHaveBodyWithWholePluginsComment)).
				Reply(201)
			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toBe(github.StatusFailure, testkeeper.NoTestsMessage, expectedContext, testkeeper.NoTestsDetailsPageName))).
				Reply(201)

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/run_cmd/trigger_run_all_comment_by_admin.json")
			issueCommentEvent := TriggerIssueCommentEvent(statusPayload, gogh.IssueCommentEvent{})

			// when
			err := handler.HandleIssueCommentEvent(log, issueCommentEvent)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should approve newly created pull request with tests when "+command.RunCommentPrefix+" "+testkeeper.ProwPluginName+" command is triggered by pr reviewer", func() {
			NonExistingRawGitHubFiles("test-keeper.yml", "test-keeper.yaml")

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/2").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/run_cmd/pr_details.json"))

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/2/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/with_tests/changes.json"))

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/issues/2/comments").
				Reply(200).
				BodyString("[]")

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/collaborators/bartoszmajsak-test/permission").
				Reply(200).
				BodyString(`{"permission": "read"}`)

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/2/reviews").
				Reply(200).
				BodyString(`[]`)

			// This way we implicitly verify that call happened after `HandleEvent` call
			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/issues/2/comments").
				SetMatcher(ExpectPayload(toHaveBodyWithWholePluginsComment)).
				Reply(201)

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toBe(github.StatusSuccess, testkeeper.TestsExistMessage, expectedContext, testkeeper.TestsExistDetailsPageName))).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/run_cmd/trigger_run_test-keeper_comment_by_pr_reviewer.json")
			issueCommentEvent := TriggerIssueCommentEvent(statusPayload, gogh.IssueCommentEvent{})

			// when
			err := handler.HandleIssueCommentEvent(log, issueCommentEvent)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should do nothing for newly created pull request with tests when "+command.RunCommentPrefix+" work-in-progress command is triggered by pr reviewer", func() {
			// given
			NonExistingRawGitHubFiles("test-keeper.yml", "test-keeper.yaml")

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/2").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/run_cmd/pr_details.json"))

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/2/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/with_tests/changes.json"))

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/issues/2/comments").
				Reply(200).
				BodyString("[]")

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/collaborators/bartoszmajsak-test/permission").
				Reply(200).
				BodyString(`{"permission": "read"}`)

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/2/reviews").
				Reply(200).
				BodyString(`[]`)

			// This way we implicitly verify that call not happened after `HandleEvent` call
			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/issues/2/comments").
				Times(0)

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				Times(0)

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/run_cmd/trigger_run_work-in-progress_comment_by_pr_reviewer.json")
			issueCommentEvent := TriggerIssueCommentEvent(statusPayload, gogh.IssueCommentEvent{})

			// when
			err := handler.HandleIssueCommentEvent(log, issueCommentEvent)

			// then
			Ω(err).ShouldNot(HaveOccurred())
		})
	})
})

func gockEmptyComments(prNumber int) {
	gock.New("https://api.github.com").
		Get(fmt.Sprintf("/repos/%s/issues/%d/comments", repositoryName, prNumber)).
		MatchParam("per_page", "100").
		MatchParam("page", "1").
		Reply(200).
		BodyString("[]")
}

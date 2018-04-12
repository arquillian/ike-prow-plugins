package testkeeper_test

import (
	"fmt"
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

const (
	repositoryName = "bartoszmajsak/wfswarm-booster-pipeline-test"
	botName        = "alien-ike"
)

var expectedContext = strings.Join([]string{botName, testkeeper.ProwPluginName}, "/")

var _ = Describe("Test Keeper Plugin features", func() {

	Context("Pull Request handling", func() {

		var handler *testkeeper.GitHubTestEventsHandler

		log := NewDiscardOutLogger()

		toBe := func(status, description, context, detailsLink string) func(statusPayload map[string]interface{}) bool {
			return func(statusPayload map[string]interface{}) bool {
				return Expect(statusPayload).To(SatisfyAll(
					HaveState(status),
					HaveDescription(description),
					HaveContext(context),
					HaveTargetURL(detailsLink),
				))
			}
		}

		toHaveBodyWithWholePluginsComment := func(statusPayload map[string]interface{}) bool {
			return Expect(statusPayload).To(SatisfyAll(
				HaveBodyThatContains(fmt.Sprintf(github.PluginTitleTemplate, testkeeper.ProwPluginName)),
				HaveBodyThatContains("@bartoszmajsak"),
			))
		}

		BeforeEach(func() {
			gock.Off()

			handler = &testkeeper.GitHubTestEventsHandler{Client: NewDefaultGitHubClient(), BotName: botName}
		})

		It("should approve opened pull request when tests included", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/2/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/with_tests/changes.json"))

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toBe(github.StatusSuccess, testkeeper.TestsExistMessage, expectedContext, testkeeper.TestsExistDetailsLink))).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/with_tests/status_opened.json")

			// when
			err := handler.HandleEvent(log, github.PullRequest, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should approve opened pull request when tests included based on configured pattern and defaults (implicitly combined)", func() {
			// given
			gock.New("https://raw.githubusercontent.com").
				Get(repositoryName + "/5d6e9b25da90edfc19f488e595e0645c081c1575/test-keeper.yml").
				Reply(200).
				BodyString("test_patterns: ['**/*_test_suite.go']\n" +
					"skip_validation_for: ['README.adoc']")

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/2/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/with_tests/changes_go_files.json"))

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toBe(github.StatusSuccess, testkeeper.TestsExistMessage, expectedContext, testkeeper.TestsExistDetailsLink))).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/with_tests/status_opened.json")

			// when
			err := handler.HandleEvent(log, github.PullRequest, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should approve new pull request without tests when it comes with configuration excluding all files from test presence check (implicitly combined)", func() {
			// given
			gock.New("https://raw.githubusercontent.com").
				Get(repositoryName + "/5d6e9b25da90edfc19f488e595e0645c081c1575/test-keeper.yml").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/test-keeper-ignore-randomfile.yml"))

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/1/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/changes-with-test-keeper-config-excluding-other-file-from-PR.json"))

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toBe(github.StatusSuccess, testkeeper.OkOnlySkippedFilesMessage, expectedContext, testkeeper.OkOnlySkippedFilesDetailsLink))).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/without_tests/status_opened.json")

			// when
			err := handler.HandleEvent(log, github.PullRequest, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should reject opened pull request when no tests are matching defined pattern with no defaults implicitly combined", func() {
			// given
			gock.New("https://raw.githubusercontent.com").
				Get(repositoryName + "/5d6e9b25da90edfc19f488e595e0645c081c1575/test-keeper.yml").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/with_tests/test-keeper.yml"))

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/2/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/with_tests/changes_go_files.json"))

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/issues/2/comments").
				Reply(200).
				BodyString("[]")

			// This way we implicitly verify that call happened after `HandleEvent` call
			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toBe(github.StatusFailure, testkeeper.NoTestsMessage, expectedContext, testkeeper.NoTestsDetailsLink))).
				Reply(201)
			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/issues/2/comments").
				SetMatcher(ExpectPayload(toHaveBodyWithWholePluginsComment)).
				Reply(201)

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/with_tests/status_opened.json")

			// when
			err := handler.HandleEvent(log, github.PullRequest, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should block newly created pull request when no tests are included", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/1/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/changes.json"))

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/issues/1/comments").
				Reply(200).
				BodyString("[]")

			// This way we implicitly verify that call happened after `HandleEvent` call
			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toBe(github.StatusFailure, testkeeper.NoTestsMessage, expectedContext, testkeeper.NoTestsDetailsLink))).
				Reply(201)
			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/issues/1/comments").
				SetMatcher(ExpectPayload(toHaveBodyWithWholePluginsComment)).
				Reply(201)

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/without_tests/status_opened.json")

			// when
			err := handler.HandleEvent(log, github.PullRequest, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should not block newly created pull request when documentation and build files are the only changes", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/1/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/build_and_docs_only_changes.json"))

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toBe(github.StatusSuccess, testkeeper.OkOnlySkippedFilesMessage, expectedContext, testkeeper.OkOnlySkippedFilesDetailsLink))).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/without_tests/status_opened.json")

			// when
			err := handler.HandleEvent(log, github.PullRequest, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should block newly created pull request when deletions in the tests are the only changes", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/1/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/deletions_only_changes_in_tests.json"))
			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/issues/1/comments").
				Reply(200).
				BodyString("[]")

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toBe(github.StatusFailure, testkeeper.NoTestsMessage, expectedContext, testkeeper.NoTestsDetailsLink))).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/issues/1/comments").
				SetMatcher(ExpectPayload(toHaveBodyWithWholePluginsComment)).
				Reply(201)

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/without_tests/status_opened.json")

			// when
			err := handler.HandleEvent(log, github.PullRequest, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should block newly created pull request when there are changes in the business logic but only deletions in the tests", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/1/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/prod_code_changes_with_deletion_only_in_tests.json"))
			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/issues/1/comments").
				Reply(200).
				BodyString("[]")

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toBe(github.StatusFailure, testkeeper.NoTestsMessage, expectedContext, testkeeper.NoTestsDetailsLink))).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/issues/1/comments").
				SetMatcher(ExpectPayload(toHaveBodyWithWholePluginsComment)).
				Reply(201)

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/without_tests/status_opened.json")

			// when
			err := handler.HandleEvent(log, github.PullRequest, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should skip test existence check when "+testkeeper.BypassCheckComment+" command is used by admin user", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/1").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/pr_details.json"))

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/collaborators/bartoszmajsak/permission").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/collaborators_repo-admin_permission.json"))

			toHaveEnforcedSuccessState := func(statusPayload map[string]interface{}) bool {
				return Expect(statusPayload).To(SatisfyAll(
					HaveState(github.StatusSuccess),
					HaveDescription(fmt.Sprintf(testkeeper.ApprovedByMessage, "bartoszmajsak")),
				))
			}

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toHaveEnforcedSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/without_tests/skip_comment_by_admin.json")

			// when
			err := handler.HandleEvent(log, github.IssueComment, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should ignore "+testkeeper.BypassCheckComment+" when used by non-admin user", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/1").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/pr_details.json"))

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/collaborators/bartoszmajsak-test/permission").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/collaborators_external-user_permission.json"))

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toBe(github.StatusFailure, testkeeper.NoTestsMessage, expectedContext, testkeeper.NoTestsDetailsLink))).
				Reply(201)

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/issues/1/comments").
				SetMatcher(
					ExpectPayload(
							ToHaveBodyContaining("Hey @bartoszmajsak-test! It seems you tried to trigger `/ok-without-tests` command"),
							ToHaveBodyContaining("You have to be admin or requested reviewer, but not pull request creator"))).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/without_tests/skip_comment_by_external.json")

			// when
			err := handler.HandleEvent(log, github.IssueComment, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})
	})
})

package plugin_test

import (
	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin"
	gogh "github.com/google/go-github/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gock "gopkg.in/h2non/gock.v1"
	keeper "github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper/plugin"
)

const (
	repositoryName = "bartoszmajsak/wfswarm-booster-pipeline-test"
)

var _ = Describe("Test Keeper Plugin features", func() {

	Context("Pull Request handling", func() {

		var handler *keeper.GitHubTestEventsHandler

		log := CreateNullLogger()

		toHaveSuccessState := func(statusPayload map[string]interface{}) bool {
			return Expect(statusPayload).To(SatisfyAll(
				HaveState("success"),
				HaveDescription("There are some tests :)"),
			))
		}

		toHaveFailureState := func(statusPayload map[string]interface{}) bool {
			return Expect(statusPayload).To(SatisfyAll(
				HaveState("failure"),
				HaveDescription("No tests in this PR :("),
			))
		}

		toHaveBodyWithWholePluginsComment := func(statusPayload map[string]interface{}) bool {
			return Expect(statusPayload).To(SatisfyAll(
				HaveBodyThatContains(fmt.Sprintf(plugin.PluginTitleTemplate, keeper.ProwPluginName)),
				HaveBodyThatContains("@bartoszmajsak"),
			))
		}

		BeforeEach(func() {
			gock.Off()

			client := gogh.NewClient(nil) // TODO with hoverfly/go-vcr we might want to use tokens instead to capture real traffic
			handler = &keeper.GitHubTestEventsHandler{Client: client}
		})

		It("should approve opened pull request when tests included", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/2/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/with_tests/changes.json"))

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toHaveSuccessState)).
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
				BodyString("test_pattern: '(__test\\.go)$'\n" +
					"skip_validation_for: 'README.adoc'")

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/2/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/with_tests/changes_go_files.json"))

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/with_tests/status_opened.json")

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
				BodyString("{}")

			// This way we implicitly verify that call happened after `HandleEvent` call
			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toHaveFailureState)).
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
				BodyString("{}")

			// This way we implicitly verify that call happened after `HandleEvent` call
			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toHaveFailureState)).
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
				SetMatcher(ExpectPayload(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/without_tests/status_opened.json")

			// when
			err := handler.HandleEvent(log, github.PullRequest, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should skip test existence check when "+keeper.SkipComment+" command is used by admin user", func() {
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
					HaveState("success"),
					HaveDescription("PR is fine without tests says @bartoszmajsak"),
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

		It("should ignore "+keeper.SkipComment+" when used by non-admin user", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/commits/5d6e9b25da90edfc19f488e595e0645c081c1575").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/pr_details.json"))

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/collaborators/bartoszmajsak-test/permission").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/collaborators_external-user_permission.json"))

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/without_tests/skip_comment_by_external.json")

			// when
			err := handler.HandleEvent(log, github.IssueComment, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})
	})
})

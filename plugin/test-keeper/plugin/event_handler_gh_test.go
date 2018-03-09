package plugin_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/arquillian/ike-prow-plugins/plugin/internal/test"
	"github.com/arquillian/ike-prow-plugins/plugin/test-keeper/plugin"
	"gopkg.in/h2non/gock.v1"
	"github.com/arquillian/ike-prow-plugins/plugin/github"
	gogh "github.com/google/go-github/github"
)

const eventGUID = "event-guid"

var _ = Describe("Test Keeper Plugin features", func() {

	Context("Pull Request handling", func() {

		var handler *plugin.GitHubTestEventsHandler

		toHaveSuccessState := func(statusPayload map[string]interface{}) (bool) {
			return Expect(statusPayload).To(SatisfyAll(
				HaveState("success"),
				HaveDescription("There are some tests :)"),
			))
		}

		toHaveFailureState := func(statusPayload map[string]interface{}) (bool) {
			return Expect(statusPayload).To(SatisfyAll(
				HaveState("failure"),
				HaveDescription("No tests in this PR :("),
			))
		}

		BeforeEach(func() {
			defer gock.Off()

			client := gogh.NewClient(nil) // TODO with hoverfly/go-vcr we might want to use tokens instead to capture real traffic
			handler = &plugin.GitHubTestEventsHandler{
				Client: client,
				Log:    CreateNullLogger(),
			}
		})

		It("should approve opened pull request when tests included", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/pulls/2/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/with_tests/changes.json"))

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectStatusCall(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/with_tests/status_opened.json")

			// when
			err := handler.HandleEvent(github.PullRequest, eventGUID, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should block newly created pull request when no tests are included", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/pulls/1/files").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/changes.json"))

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectStatusCall(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/without_tests/status_opened.json")

			// when
			err := handler.HandleEvent(github.PullRequest, eventGUID, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should skip test existence check when "+plugin.SkipComment+" command is used by admin user", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/pulls/1").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/pr_details.json"))

			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/collaborators/bartoszmajsak/permission").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/collaborators_repo-admin_permission.json"))

			toHaveEnforcedSuccessState := func(statusPayload map[string]interface{}) (bool) {
				return Expect(statusPayload).To(SatisfyAll(
					HaveState("success"),
					HaveDescription("PR is fine without tests says @bartoszmajsak"),
				))
			}

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectStatusCall(toHaveEnforcedSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/without_tests/skip_comment_by_admin.json")

			// when
			err := handler.HandleEvent(github.IssueComment, eventGUID, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should ignore "+plugin.SkipComment+" when used by non-admin user", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/commits/5d6e9b25da90edfc19f488e595e0645c081c1575").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/prs/without_tests/pr_details.json"))

			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/collaborators/bartoszmajsak-test/permission").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/collaborators_external-user_permission.json"))

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectStatusCall(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/prs/without_tests/skip_comment_by_external.json")

			// when
			err := handler.HandleEvent(github.IssueComment, eventGUID, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

	})

})

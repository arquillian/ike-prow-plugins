package wip_test

import (
	"strings"

	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/work-in-progress"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"

	"github.com/arquillian/ike-prow-plugins/pkg/command"
	"github.com/arquillian/ike-prow-plugins/pkg/github"
)

const (
	botName        = "alien-ike"
	repositoryName = "bartoszmajsak/wfswarm-booster-pipeline-test"
)

var expectedContext = strings.Join([]string{botName, wip.ProwPluginName}, "/")

var _ = Describe("Test Keeper Plugin features", func() {

	Context("Pull Request title change trigger", func() {

		var handler *wip.GitHubWIPPRHandler

		log := log.NewTestLogger()

		toHaveSuccessState := SoftlySatisfyAll(
			HaveState(github.StatusSuccess),
			HaveDescription(wip.ReadyForReviewMessage),
			HaveContext(expectedContext),
			HaveTargetURL(wip.ReadyForReviewDetailsLink),
		)

		toHaveFailureState := SoftlySatisfyAll(
			HaveState(github.StatusFailure),
			HaveDescription(wip.InProgressMessage),
			HaveContext(expectedContext),
			HaveTargetURL(wip.InProgressDetailsLink),
		)

		BeforeEach(func() {
			defer gock.OffAll()
			handler = &wip.GitHubWIPPRHandler{Client: NewDefaultGitHubClient(), BotName: botName}
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should mark opened PR as ready for review if not prefixed with WIP", func() {
			// given
			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectPayload(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/ready_pr_opened.json")

			// when
			err := handler.HandleEvent(log, github.PullRequest, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark opened PR as work-in-progress when prefixed with WIP", func() {
			// given
			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectPayload(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/wip_pr_opened.json")

			// when
			err := handler.HandleEvent(log, github.PullRequest, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark status as failed (thus block PR merge) when title updated to contain WIP", func() {
			// given
			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectPayload(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/pr_edited_wip_added.json")

			// when
			err := handler.HandleEvent(log, github.PullRequest, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())

		})

		It("should mark status as success (thus unblock PR merge) when title has WIP removed", func() {
			// given
			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectPayload(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/pr_edited_wip_removed.json")

			// when
			err := handler.HandleEvent(log, github.PullRequest, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark opened PR as ready for review if not prefixed with WIP when "+command.RunCommentPrefix+" all command is used by external user", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/11").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls//pr_details.json"))

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/11/reviews").
				Reply(200).
				BodyString(`[]`)

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/collaborators/bartoszmajsak-test/permission").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/collaborators_external-user_permission.json"))

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectPayload(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/run_comment_pr_by_external.json")

			// when
			err := handler.HandleEvent(log, github.IssueComment, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark opened PR as work-in-progress if prefixed with WIP when "+command.RunCommentPrefix+" all command is used by admin", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/11").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls//pr_details_wip.json"))

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/11/reviews").
				Reply(200).
				BodyString(`[]`)

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/collaborators/bartoszmajsak/permission").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/collaborators_repo-admin_permission.json"))

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectPayload(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/run_comment_wip_pr_by_pr_creator.json")

			// when
			err := handler.HandleEvent(log, github.IssueComment, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

	})

})

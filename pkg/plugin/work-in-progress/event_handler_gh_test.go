package wip_test

import (
	"strings"

	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/work-in-progress"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"

	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/command"
	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper"
)

const (
	botName        = "alien-ike"
	repositoryName = "bartoszmajsak/wfswarm-booster-pipeline-test"
)

var (
	expectedContext = strings.Join([]string{botName, wip.ProwPluginName}, "/")
	docStatusRoot   = fmt.Sprintf("%s/status/%s", plugin.DocumentationURL, wip.ProwPluginName)
)

var _ = Describe("Work In Progress Plugin features", func() {

	var handler *wip.GitHubWIPPRHandler

	log := log.NewTestLogger()
	configFilePath := ghservice.ConfigHome + wip.ProwPluginName

	toHaveSuccessState := SoftlySatisfyAll(
		HaveState(github.StatusSuccess),
		HaveDescription(wip.ReadyForReviewMessage),
		HaveContext(expectedContext),
		HaveTargetURL(fmt.Sprintf("%s/%s/%s.html", docStatusRoot, "success", wip.ReadyForReviewDetailsPageName)),
	)

	toHaveFailureState := SoftlySatisfyAll(
		HaveState(github.StatusFailure),
		HaveDescription(wip.InProgressMessage),
		HaveContext(expectedContext),
		HaveTargetURL(fmt.Sprintf("%s/%s/%s.html", docStatusRoot, "failure", wip.InProgressDetailsPageName)),
	)

	Context("Pull Request label change trigger", func() {
		BeforeEach(func() {
			defer gock.OffAll()
			handler = &wip.GitHubWIPPRHandler{Client: NewDefaultGitHubClient(), BotName: botName}
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should mark opened PR as work-in-progress when labeled with WIP", func() {
			// given
			NonExistingRawGitHubFiles("work-in-progress.yml", "work-in-progress.yaml")

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectPayload(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/pr_labeled_wip.json")

			// when
			err := handler.HandleEvent(log, github.PullRequest, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark opened PR as ready for review if WIP label removed", func() {
			// given
			NonExistingRawGitHubFiles("work-in-progress.yml", "work-in-progress.yaml")

			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/issues/4/labels").
				Reply(200).
				BodyString("[]")

			gock.New("https://api.github.com").
				Patch("repos/bartoszmajsak/wfswarm-booster-pipeline-test/pulls/4").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/pr_edited_with_title_change.json"))

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectPayload(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/wip_pr_unlabeled.json")

			// when
			err := handler.HandleEvent(log, github.PullRequest, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Context("Pull Request title change trigger", func() {
		BeforeEach(func() {
			defer gock.OffAll()
			handler = &wip.GitHubWIPPRHandler{Client: NewDefaultGitHubClient(), BotName: botName}
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should mark opened PR as ready for review if not prefixed with WIP", func() {
			// given
			NonExistingRawGitHubFiles("work-in-progress.yml", "work-in-progress.yaml")

			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/issues/4/labels").
				Reply(200).
				BodyString("[]")

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
			NonExistingRawGitHubFiles("work-in-progress.yml", "work-in-progress.yaml")

			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/issues/4/labels").
				Reply(200).
				BodyString("[]")

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/issues/4/labels").
				SetMatcher(ExpectPayload(To(HaveBodyThatContains("work-in-progress")))).
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/wip_pr_created_with_label.json"))

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

		It("should mark opened PR as work-in-progress when title starts with configured prefix", func() {
			// given
			gock.New("https://raw.githubusercontent.com").
				Get("bartoszmajsak/wfswarm-booster-pipeline-test/8111c2d99b596877ff8e2059409688d83487da0e/" + configFilePath + ".yml").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/work-in-progress.yml"))

			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/issues/4/labels").
				Reply(200).
				BodyString("[]")

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/issues/4/labels").
				SetMatcher(ExpectPayload(To(HaveBodyThatContains("wip")))).
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/wip_pr_created_with_label.json"))

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectPayload(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/custom_prefix_pr_opened.json")

			// when
			err := handler.HandleEvent(log, github.PullRequest, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark status as failed (thus block PR merge) when title updated to contain WIP", func() {
			// given
			NonExistingRawGitHubFiles("work-in-progress.yml", "work-in-progress.yaml")

			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/issues/4/labels").
				Reply(200).
				BodyString("[]")

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/issues/4/labels").
				SetMatcher(ExpectPayload(To(HaveBodyThatContains("work-in-progress")))).
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/pr_edited_with_label.json"))

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
			NonExistingRawGitHubFiles("work-in-progress.yml", "work-in-progress.yaml")

			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/issues/4/labels").
				Reply(200).
				BodyString(`[{"id": 934813958,` +
					`"url": "https://api.github.com/repos/bartoszmajsak/wfswarm-booster-pipeline-test/labels/work-in-progress",` +
					`"name": "work-in-progress", "color": "ededed", "default": false}]`)

			gock.New("https://api.github.com").
				Delete("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/issues/4/labels/work-in-progress").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/pr_edited_with_unlabel.json"))

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

	})

	Context("Trigger work-in-progress plugin by triggering comment on pull request", func() {
		BeforeEach(func() {
			defer gock.OffAll()
			handler = &wip.GitHubWIPPRHandler{Client: NewDefaultGitHubClient(), BotName: botName}
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should mark opened PR as ready for review if not prefixed with WIP when "+command.RunCommentPrefix+" "+wip.ProwPluginName+" command is triggered by pr creator", func() {
			// given
			NonExistingRawGitHubFiles("work-in-progress.yml", "work-in-progress.yaml")

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/11").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/pr_details.json"))

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/11/reviews").
				Reply(200).
				BodyString(`[]`)

			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/issues/11/labels").
				Reply(200).
				BodyString("[]")

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/collaborators/bartoszmajsak-test/permission").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/collaborators_external-user_permission.json"))

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectPayload(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/trigger_run_work-in-progress_on_pr_by_pr_creator.json")

			// when
			err := handler.HandleEvent(log, github.IssueComment, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark opened PR as work-in-progress if prefixed with WIP when "+command.RunCommentPrefix+" all command is used by admin", func() {
			// given
			NonExistingRawGitHubFiles("work-in-progress.yml", "work-in-progress.yaml")

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
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/issues/11/labels").
				Reply(200).
				BodyString("[]")

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/issues/11/labels").
				SetMatcher(ExpectPayload(To(HaveBodyThatContains("work-in-progress")))).
				Reply(200).
				BodyString("work-in-progress")

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectPayload(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/trigger_run_all_on_wip_pr_by_admin.json")

			// when
			err := handler.HandleEvent(log, github.IssueComment, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should approve newly created pull request with tests when "+command.RunCommentPrefix+" "+testkeeper.ProwPluginName+" command is triggered by pr creator", func() {
			// given
			NonExistingRawGitHubFiles("work-in-progress.yml", "work-in-progress.yaml")

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/11").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/pr_details.json"))

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
				Times(0) // This way we implicitly verify that call not happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/trigger_run_test-keeper_on_pr_by_pr_creator.json")

			// when
			err := handler.HandleEvent(log, github.IssueComment, statusPayload)

			// then
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

})

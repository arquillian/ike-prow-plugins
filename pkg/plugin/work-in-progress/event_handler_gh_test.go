package wip_test

import (
	"strings"

	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/work-in-progress"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
)

const (
	botName = "alien-ike"
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
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/wip_pr_created_with_label.json"))

			// This way we implicitly verify that call happened after `HandleEvent` call
			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectPayload(toHaveFailureState)).
				Reply(201)

			statusPayload := LoadFromFile("test_fixtures/github_calls/wip_pr_opened.json")

			// when
			err := handler.HandleEvent(log, github.PullRequest, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark opened PR as work-in-progress when title starts with configured prefix", func() {
			// given
			gock.New("https://raw.githubusercontent.com").
				Get("bartoszmajsak/wfswarm-booster-pipeline-test/8111c2d99b596877ff8e2059409688d83487da0e/work-in-progress.yml").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/work-in-progress.yml"))

			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/issues/4/labels").
				Reply(200).
				BodyString("[]")

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/issues/4/labels").
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
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/pr_edited_with_label.json"))

			// This way we implicitly verify that call happened after `HandleEvent` call
			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectPayload(toHaveFailureState)).
				Reply(201)

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

			// This way we implicitly verify that call happened after `HandleEvent` call
			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectPayload(toHaveSuccessState)).
				Reply(201)

			statusPayload := LoadFromFile("test_fixtures/github_calls/pr_edited_wip_removed.json")

			// when
			err := handler.HandleEvent(log, github.PullRequest, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

	})

})

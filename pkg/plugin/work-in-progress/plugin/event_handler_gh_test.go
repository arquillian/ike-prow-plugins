package plugin

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"gopkg.in/h2non/gock.v1"
	gogh "github.com/google/go-github/github"
	"github.com/arquillian/ike-prow-plugins/pkg/github"
)

var _ = Describe("Test Keeper Plugin features", func() {

Context("Pull Request title change trigger", func() {

		var handler *GitHubWIPPRHandler

		toHaveSuccessState := func(statusPayload map[string]interface{}) (bool) {
			return Expect(statusPayload).To(SatisfyAll(
				HaveState("success"),
				HaveDescription("PR is ready for review and merge"),
			))
		}

		toHaveFailureState := func(statusPayload map[string]interface{}) (bool) {
			return Expect(statusPayload).To(SatisfyAll(
				HaveState("failure"),
				HaveDescription("PR is in progress and can't be merged yet. You might want to wait with review as well"),
			))
		}

		BeforeEach(func() {
			defer gock.Off()
			client := gogh.NewClient(nil) // TODO with hoverfly/go-vcr we might want to use tokens instead to capture real traffic
			handler = &GitHubWIPPRHandler{
				Client: client,
				Log:    CreateNullLogger(),
			}
		})

		It("should mark opened PR as ready for review if not prefixed with WIP", func() {
			// given
			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectStatusCall(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/ready_pr_opened.json")

			// when
			err := handler.HandleEvent(github.PullRequest, "random", statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark opened PR as work-in-progress when prefixed with WIP", func() {
			// given
			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectStatusCall(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/wip_pr_opened.json")

			// when
			err := handler.HandleEvent(github.PullRequest, "random", statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark status as failed (thus block PR merge) when title updated to contain WIP", func() {
			// given
			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectStatusCall(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/pr_edited_wip_added.json")

			// when
			err := handler.HandleEvent(github.PullRequest, "random", statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())

		})

		It("should mark status as success (thus unblock PR merge) when title has WIP removed", func() {
			// given
			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectStatusCall(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/pr_edited_wip_removed.json")

			// when
			err := handler.HandleEvent(github.PullRequest, "random", statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

	})

})

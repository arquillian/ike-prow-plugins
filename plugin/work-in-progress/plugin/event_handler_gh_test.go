package plugin_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/arquillian/ike-prow-plugins/plugin/internal/test"

	"github.com/arquillian/ike-prow-plugins/plugin/work-in-progress/plugin"
	"gopkg.in/h2non/gock.v1"
	"github.com/sirupsen/logrus"
	"github.com/google/go-github/github"
	"github.com/arquillian/ike-prow-plugins/plugin/github"
	"io/ioutil"
)

var _ = Describe("Test Keeper Plugin features", func() {
	Context("Pull Request title change trigger", func() {

		var handler *plugin.GitHubWIPPRHandler

		BeforeEach(func() {
			defer gock.Off()

			nullLogger := logrus.New()
			nullLogger.Out = ioutil.Discard
			logger := logrus.NewEntry(nullLogger)

			client := github.NewClient(nil) // TODO with hoverfly/go-vcr we might want to use tokens instead to capture real traffic
			handler = &plugin.GitHubWIPPRHandler{
				Client: client,
				Log:    logger,
			}
		})

		It("should mark opened PR as ready for review if not prefixed with WIP", func() {
			// given
			toHaveSuccessState := func(statusPayload map[string]interface{}) (bool) {
				return Expect(statusPayload["state"]).To(Equal("success"))
			}

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectStatusCall(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			// when
			handler.HandleEvent(githubevents.PullRequest, "random", EventPayload("test_fixtures/github_calls/ready_pr_opened.json"))

			// then - implicit verification of /statuses call occurrence with proper payload
		})

		It("should mark opened PR as work-in-progress when prefixed with WIP", func() {
			// given
			toHaveFailureState := func(statusPayload map[string]interface{}) (bool) {
				return Expect(statusPayload["state"]).To(Equal("failure"))
			}

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectStatusCall(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			// when
			handler.HandleEvent(githubevents.PullRequest, "random", EventPayload("test_fixtures/github_calls/wip_pr_opened.json"))

			// then - implicit verification of /statuses call occurrence with proper payload
		})

		It("should mark status as failed (thus block PR merge) when title updated to contain WIP", func() {
			// given
			toHaveFailureState := func(statusPayload map[string]interface{}) (bool) {
				return Expect(statusPayload["state"]).To(Equal("failure"))
			}

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectStatusCall(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			// when
			handler.HandleEvent(githubevents.PullRequest, "random", EventPayload("test_fixtures/github_calls/pr_edited_wip_added.json"))

			// then - implicit verification of /statuses call occurrence with proper payload

		})

		It("should mark status as success (thus unblock PR merge) when title has WIP removed", func() {
			// given
			toHaveSuccessState := func(statusPayload map[string]interface{}) (bool) {
				return Expect(statusPayload["state"]).To(Equal("success"))
			}

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectStatusCall(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			// when
			handler.HandleEvent(githubevents.PullRequest, "random", EventPayload("test_fixtures/github_calls/pr_edited_wip_removed.json"))

			// then - implicit verification of /statuses call occurrence with proper payload
		})

	})

})

package plugin_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/arquillian/ike-prow-plugins/plugin/test-keeper/plugin"
	"gopkg.in/h2non/gock.v1"
	"github.com/sirupsen/logrus"
	"github.com/arquillian/ike-prow-plugins/plugin/github"
	"os"
	"io/ioutil"
	"fmt"
	"github.com/google/go-github/github"
	"net/http"
	"encoding/json"
	"io"
)

const eventGUID = "event-guid"

var _ = Describe("Test Keeper Plugin features", func() {

	var logger *logrus.Entry

	BeforeSuite(func() {
		nullLogger := logrus.New()
		nullLogger.Out = ioutil.Discard
		logger = logrus.NewEntry(nullLogger)
	})

	Context("Pull Request handling", func() {

		var handler *plugin.GitHubTestEventsHandler

		BeforeEach(func() {
			defer gock.Off()

			client := github.NewClient(nil) // TODO with hoverfly/go-vcr we might want to use tokens instead to capture real traffic
			handler = &plugin.GitHubTestEventsHandler{
				Client: client,
				Log:    logger,
			}
		})

		It("should approve opened pull request when tests included", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/pulls/2/files").
				Reply(200).
				Body(fromJson("test_fixtures/github_calls/pr_2_with_tests.json"))

			toHaveSuccessState := func(statusPayload map[string]interface{}) (bool) {
				return Expect(statusPayload["state"]).To(Equal("success"))
			}

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(expectStatusCall(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			// when
			handler.HandleEvent(githubevents.PullRequest, eventGUID, eventPayload("test_fixtures/github_calls/pr_2_opened_status.json"))

			// then - implicit verification of /statuses call occurrence with proper payload
		})

		It("should block newly created pull request when no tests are included", func() {

			// given
			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/pulls/1/files").
				Reply(200).
				Body(fromJson("test_fixtures/github_calls/pr_1_without_tests.json"))


			toHaveFailureState := func(statusPayload map[string]interface{}) (bool) {
				return Expect(statusPayload["state"]).To(Equal("failure"))
			}

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(expectStatusCall(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			// when
			handler.HandleEvent(githubevents.PullRequest, eventGUID, eventPayload("test_fixtures/github_calls/pr_1_opened_status.json"))

			// then - implicit verification of /statuses call occurrence with proper payload
		})

		It("should skip test existence check when "+plugin.SkipComment+" command is used by admin user", func() {

			// given
			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/pulls/1").
				Reply(200).
				Body(fromJson("test_fixtures/github_calls/pr_1_content.json"))

			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/collaborators/bartoszmajsak/permission").
				Reply(200).
				Body(fromJson("test_fixtures/github_calls/collaborators_bartoszmajsak_permission.json"))

			toHaveSuccessState := func(statusPayload map[string]interface{}) (bool) {
				return Expect(statusPayload["state"]).To(Equal("success"))
			}

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(expectStatusCall(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			// when
			handler.HandleEvent(githubevents.IssueComment, eventGUID, eventPayload("test_fixtures/github_calls/pr_1_skip_comment_by_admin.json"))

			// then - implicit verification of /statuses call occurrence with proper payload
		})

		It("should ignore "+plugin.SkipComment+" when used by non-admin user", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/pulls/1").
				Reply(200).
				Body(fromJson("test_fixtures/github_calls/pr_1_content.json"))

			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/collaborators/bartoszmajsak-test/permission").
				Reply(200).
				Body(fromJson("test_fixtures/github_calls/collaborators_bartoszmajsak-test_permission.json"))

			toHaveFailureState := func(statusPayload map[string]interface{}) (bool) {
				return Expect(statusPayload["state"]).To(Equal("failure"))
			}

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(expectStatusCall(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			// when
			handler.HandleEvent(githubevents.IssueComment, eventGUID, eventPayload("test_fixtures/github_calls/pr_1_skip_comment_by_external.json"))

			// then - implicit verification of /statuses call occurrence with proper payload
		})

	})

})

func eventPayload(payloadFile string) []byte {
	payload, err := ioutil.ReadFile(payloadFile)
	if err != nil {
		Fail(fmt.Sprintf("Unable to load test fixture. Reason: %q", err))
	}
	return payload
}

func fromJson(filePath string) io.Reader {
	file, err := os.Open(filePath)
	if err != nil {
		Fail(fmt.Sprintf("Unable to load test fixture. Reason: %q", err))
	}

	return file
}

func expectStatusCall(payloadAssert func(statusPayload map[string]interface{}) (bool)) gock.Matcher {
	matcher := gock.NewBasicMatcher()
	matcher.Add(func(req *http.Request, _ *gock.Request) (bool, error) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return false, err
		}
		var statusPayload map[string]interface{}
		err = json.Unmarshal(body, &statusPayload)
		payloadExpectations := payloadAssert(statusPayload)
		return payloadExpectations, err
	})
	return matcher
}

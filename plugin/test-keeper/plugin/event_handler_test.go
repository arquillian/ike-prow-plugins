package plugin_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/arquillian/ike-prow-plugins/plugin/test-keeper/plugin"
	"gopkg.in/h2non/gock.v1"
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

	Context("Pull Request handling", func() {

		var handler *plugin.GitHubTestEventsHandler

		BeforeEach(func() {
			defer gock.Off()

			client := github.NewClient(nil) // TODO with hoverfly/go-vcr we might want to use tokens instead to capture real traffic
			handler = &plugin.GitHubTestEventsHandler{
				Client: client,
				Log:    Logger,
			}
		})

		It("should approve opened pull request when tests included", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/commits/5d6e9b25da90edfc19f488e595e0645c081c1575").
				Reply(200).
				Body(fromJson("test_fixtures/github_calls/prs/with_tests/changes.json"))

			toHaveSuccessState := func(statusPayload map[string]interface{}) (bool) {
				return Expect(statusPayload["state"]).To(Equal("success"))
			}

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(expectStatusCall(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			// when
			err := handler.HandleEvent(githubevents.PullRequest, eventGUID, eventPayload("test_fixtures/github_calls/prs/with_tests/status_opened.json"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Expect(err).To(BeNil())
		})

		It("should block newly created pull request when no tests are included", func() {

			// given
			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/commits/5d6e9b25da90edfc19f488e595e0645c081c1575").
				Reply(200).
				Body(fromJson("test_fixtures/github_calls/prs/without_tests/changes.json"))

			toHaveFailureState := func(statusPayload map[string]interface{}) (bool) {
				return Expect(statusPayload["state"]).To(Equal("failure"))
			}

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(expectStatusCall(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			// when
			err := handler.HandleEvent(githubevents.PullRequest, eventGUID, eventPayload("test_fixtures/github_calls/prs/without_tests/status_opened.json"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Expect(err).To(BeNil())
		})

		It("should skip test existence check when "+plugin.SkipComment+" command is used by admin user", func() {

			// given
			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/pulls/1").
				Reply(200).
				Body(fromJson("test_fixtures/github_calls/prs/without_tests/pr_details.json"))

			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/collaborators/bartoszmajsak/permission").
				Reply(200).
				Body(fromJson("test_fixtures/github_calls/collaborators_repo-admin_permission.json"))

			toHaveSuccessState := func(statusPayload map[string]interface{}) (bool) {
				return Expect(statusPayload["state"]).To(Equal("success"))
			}

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(expectStatusCall(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			// when
			err := handler.HandleEvent(githubevents.IssueComment, eventGUID, eventPayload("test_fixtures/github_calls/prs/without_tests/skip_comment_by_admin.json"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Expect(err).To(BeNil())
		})

		It("should ignore "+plugin.SkipComment+" when used by non-admin user", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/commits/5d6e9b25da90edfc19f488e595e0645c081c1575").
				Reply(200).
				Body(fromJson("test_fixtures/github_calls/prs/without_tests/pr_details.json"))

			gock.New("https://api.github.com").
				Get("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/collaborators/bartoszmajsak-test/permission").
				Reply(200).
				Body(fromJson("test_fixtures/github_calls/collaborators_external-user_permission.json"))

			toHaveFailureState := func(statusPayload map[string]interface{}) (bool) {
				return Expect(statusPayload["state"]).To(Equal("failure"))
			}

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(expectStatusCall(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			// when
			err := handler.HandleEvent(githubevents.IssueComment, eventGUID, eventPayload("test_fixtures/github_calls/prs/without_tests/skip_comment_by_external.json"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Expect(err).To(BeNil())
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

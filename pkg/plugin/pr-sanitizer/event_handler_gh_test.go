package prsanitizer_test

import (
	"fmt"
	"strings"

	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/pr-sanitizer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
)

const (
	botName = "alien-ike"
)

var (
	expectedContext = strings.Join([]string{botName, prsanitizer.ProwPluginName}, "/")
	docStatusRoot   = fmt.Sprintf("%s/status/%s", plugin.DocumentationURL, prsanitizer.ProwPluginName)
)

var _ = Describe("PR Sanitizer Plugin features", func() {

	Context("Pull Request title change trigger", func() {

		var handler *prsanitizer.GitHubLabelsEventsHandler

		log := log.NewTestLogger()
		configFilePath := ghservice.ConfigHome + prsanitizer.ProwPluginName

		toHaveSuccessState := SoftlySatisfyAll(
			HaveState(github.StatusSuccess),
			HaveDescription(prsanitizer.SuccessMessage),
			HaveContext(expectedContext),
			HaveTargetURL(fmt.Sprintf("%s/%s/%s.html", docStatusRoot, "success", prsanitizer.SuccessDetailsLink)),
		)

		toHaveFailureState := SoftlySatisfyAll(
			HaveState(github.StatusFailure),
			HaveDescription(prsanitizer.FailureMessage),
			HaveContext(expectedContext),
			HaveTargetURL(fmt.Sprintf("%s/%s/%s.html", docStatusRoot, "failure", prsanitizer.FailureDetailsLink)),
		)

		BeforeEach(func() {
			defer gock.OffAll()
			handler = &prsanitizer.GitHubLabelsEventsHandler{Client: NewDefaultGitHubClient(), BotName: botName}
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should mark status as success if PR title prefixed with semantic commit message type", func() {
			// given
			NonExistingRawGitHubFiles("pr-sanitizer.yml", "pr-sanitizer.yaml")

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectPayload(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/semantically_correct_pr_opened.json")

			// when
			err := handler.HandleEvent(log, github.PullRequest, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark status as failed (thus block PR merge) when not prefixed with semantic commit message type", func() {
			// given
			NonExistingRawGitHubFiles("pr-sanitizer.yml", "pr-sanitizer.yaml")

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectPayload(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/semantically_incorrect_pr_opened.json")

			// when
			err := handler.HandleEvent(log, github.PullRequest, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark status as success when title starts with configured semantic commit message type", func() {
			// given
			gock.New("https://raw.githubusercontent.com").
				Get("bartoszmajsak/wfswarm-booster-pipeline-test/8111c2d99b596877ff8e2059409688d83487da0e/" + configFilePath + ".yml").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/pr-sanitizer.yml"))

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectPayload(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/custom_prefix_pr_edited.json")

			// when
			err := handler.HandleEvent(log, github.PullRequest, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark status as success (thus unblock PR merge) when title updated to contain semantic commit message type", func() {
			// given
			NonExistingRawGitHubFiles("pr-sanitizer.yml", "pr-sanitizer.yaml")

			gock.New("https://api.github.com").
				Post("/repos/bartoszmajsak/wfswarm-booster-pipeline-test/statuses").
				SetMatcher(ExpectPayload(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			statusPayload := LoadFromFile("test_fixtures/github_calls/pr_edited_type_added.json")

			// when
			err := handler.HandleEvent(log, github.PullRequest, statusPayload)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())

		})

	})

})

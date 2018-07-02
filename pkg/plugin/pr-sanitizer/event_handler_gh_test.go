package prsanitizer_test

import (
	"fmt"
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/command"
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
	botName        = "alien-ike"
	repositoryName = "bartoszmajsak/wfswarm-booster-pipeline-test"
)

var (
	expectedContext = strings.Join([]string{botName, prsanitizer.ProwPluginName}, "/")
	docStatusRoot   = fmt.Sprintf("%s/status/%s", plugin.DocumentationURL, prsanitizer.ProwPluginName)
)

var _ = Describe("PR Sanitizer Plugin features", func() {

	var handler *prsanitizer.GitHubPRSanitizerEventsHandler

	log := log.NewTestLogger()
	configFilePath := ghservice.ConfigHome + prsanitizer.ProwPluginName

	toHaveSuccessState := SoftlySatisfyAll(
		HaveState(github.StatusSuccess),
		HaveDescription(prsanitizer.SuccessMessage),
		HaveContext(expectedContext),
		HaveTargetURL(fmt.Sprintf("%s/%s/%s.html", docStatusRoot, "success", prsanitizer.SuccessDetailsPageName)),
	)

	toHaveFailureState := SoftlySatisfyAll(
		HaveState(github.StatusFailure),
		HaveDescription(prsanitizer.FailureMessage),
		HaveContext(expectedContext),
		HaveTargetURL(fmt.Sprintf("%s/%s/%s.html", docStatusRoot, "failure", prsanitizer.FailureDetailsPageName)),
	)

	Context("Pull Request title change trigger", func() {
		BeforeEach(func() {
			defer gock.OffAll()
			handler = &prsanitizer.GitHubPRSanitizerEventsHandler{Client: NewDefaultGitHubClient(), BotName: botName}
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should mark status as success if PR title prefixed with semantic commit message type", func() {
			// given
			NonExistingRawGitHubFiles("pr-sanitizer.yml", "pr-sanitizer.yaml")

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			event := LoadPullRequestEvent("test_fixtures/github_calls/semantically_correct_pr_opened.json")

			// when
			err := handler.HandlePullRequestEvent(log, event)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark status as failed (thus block PR merge) when not prefixed with semantic commit message type", func() {
			// given
			NonExistingRawGitHubFiles("pr-sanitizer.yml", "pr-sanitizer.yaml", "work-in-progress.yml", "work-in-progress.yaml")

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/issues/4/comments").
				SetMatcher(ExpectPayload(HaveBodyThatContains(fmt.Sprintf("Hey @%s! %s", "bartoszmajsak", prsanitizer.TitleFailureMessage)))).
				Reply(201)

			event := LoadPullRequestEvent("test_fixtures/github_calls/semantically_incorrect_pr_opened.json")

			// when
			err := handler.HandlePullRequestEvent(log, event)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark status as success when title starts with configured semantic commit message type", func() {
			// given
			gock.New("https://raw.githubusercontent.com").
				Get(repositoryName + "/8111c2d99b596877ff8e2059409688d83487da0e/" + configFilePath + ".yml").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/pr-sanitizer.yml"))

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			event := LoadPullRequestEvent("test_fixtures/github_calls/custom_prefix_pr_edited.json")

			// when
			err := handler.HandlePullRequestEvent(log, event)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark status as success (thus unblock PR merge) when title updated to contain semantic commit message type", func() {
			// given
			NonExistingRawGitHubFiles("pr-sanitizer.yml", "pr-sanitizer.yaml")

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			event := LoadPullRequestEvent("test_fixtures/github_calls/pr_edited_type_added.json")

			// when
			err := handler.HandlePullRequestEvent(log, event)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())

		})

		It("should mark status as success if PR title prefixed with wip and conforms with semantic commit message type", func() {
			// given
			NonExistingRawGitHubFiles("pr-sanitizer.yml", "pr-sanitizer.yaml", "work-in-progress.yml", "work-in-progress.yaml")

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			event := LoadPullRequestEvent("test_fixtures/github_calls/semantically_correct_wip_pr_opened.json")

			// when
			err := handler.HandlePullRequestEvent(log, event)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark status as failed (thus block PR merge) when prefixed with wip and does not conform with semantic commit message type", func() {
			// given
			NonExistingRawGitHubFiles("pr-sanitizer.yml", "pr-sanitizer.yaml", "work-in-progress.yml", "work-in-progress.yaml")

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/issues/4/comments").
				SetMatcher(ExpectPayload(HaveBodyThatContains(fmt.Sprintf("Hey @%s! %s", "bartoszmajsak", prsanitizer.TitleFailureMessage)))).
				Reply(201)

			event := LoadPullRequestEvent("test_fixtures/github_calls/semantically_incorrect_wip_pr_opened.json")

			// when
			err := handler.HandlePullRequestEvent(log, event)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

	})

	Context("Pull Request Description verifier", func() {

		toHaveFailureState = SoftlySatisfyAll(
			HaveState(github.StatusFailure),
			HaveContext(expectedContext),
			HaveTargetURL(fmt.Sprintf("%s/%s/%s.html", docStatusRoot, "failure", prsanitizer.FailureDetailsPageName)),
		)

		descriptionContentLengthMessage := fmt.Sprintf(prsanitizer.DescriptionLengthShortMessage, 50)

		BeforeEach(func() {
			defer gock.OffAll()
			handler = &prsanitizer.GitHubPRSanitizerEventsHandler{Client: NewDefaultGitHubClient(), BotName: botName}
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should mark status as failed (thus block PR merge) when PR doesn't have issue linked in the description", func() {
			// given
			NonExistingRawGitHubFiles("pr-sanitizer.yml", "pr-sanitizer.yaml")

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/issues/4/comments").
				SetMatcher(ExpectPayload(HaveBodyThatContains(fmt.Sprintf("Hey @%s! %s", "bartoszmajsak", prsanitizer.IssueLinkMissingMessage)))).
				Reply(201)

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			event := LoadPullRequestEvent("test_fixtures/github_calls/issue_link_missing_pr_opened.json")

			// when
			err := handler.HandlePullRequestEvent(log, event)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark status as failed (thus block PR merge) when PR doesn't have description", func() {
			// given
			NonExistingRawGitHubFiles("pr-sanitizer.yml", "pr-sanitizer.yaml", "work-in-progress.yml", "work-in-progress.yaml")

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/issues/4/comments").
				SetMatcher(ExpectPayload(HaveBodyThatContains(fmt.Sprintf("Hey @%s! %s %s", "bartoszmajsak", prsanitizer.TitleFailureMessage, descriptionContentLengthMessage)))).
				Reply(201)

			event := LoadPullRequestEvent("test_fixtures/github_calls/semantically_incorrect_title_missing_description_pr_opened.json")

			// when
			err := handler.HandlePullRequestEvent(log, event)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark status as failed (thus block PR merge) when PR doesn't have description length as per configuration", func() {
			// given
			gock.New("https://raw.githubusercontent.com").
				Get(repositoryName + "/6582335bd87edd6b1fa32e32d566fbdf6c2fa579/" + configFilePath + ".yml").
				Reply(200).
				BodyString("description_content_length: 100")

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/issues/4/comments").
				SetMatcher(ExpectPayload(HaveBodyThatContains(fmt.Sprintf("Hey @%s! %s", "bartoszmajsak", fmt.Sprintf(prsanitizer.DescriptionLengthShortMessage, 100))))).
				Reply(201)

			event := LoadPullRequestEvent("test_fixtures/github_calls/short_description_as_per_config_pr_opened.json")

			// when
			err := handler.HandlePullRequestEvent(log, event)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark status as failed (thus block PR merge) when prefixed with wip and does not conform with semantic commit message type and short description", func() {
			// given
			NonExistingRawGitHubFiles("pr-sanitizer.yml", "pr-sanitizer.yaml", "work-in-progress.yml", "work-in-progress.yaml")

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/issues/4/comments").
				SetMatcher(ExpectPayload(HaveBodyThatContains(fmt.Sprintf("Hey @%s! %s %s", "bartoszmajsak", prsanitizer.TitleFailureMessage, descriptionContentLengthMessage)))).
				Reply(201)

			event := LoadPullRequestEvent("test_fixtures/github_calls/semantically_incorrect_wip_short_desc_pr_opened.json")

			// when
			err := handler.HandlePullRequestEvent(log, event)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

	})

	Context("Trigger pr-sanitizer plugin by triggering comment on pull request", func() {
		BeforeEach(func() {
			defer gock.OffAll()
			handler = &prsanitizer.GitHubPRSanitizerEventsHandler{Client: NewDefaultGitHubClient(), BotName: botName}
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should mark status as success if PR title prefixed with semantic commit message type when "+command.RunCommentPrefix+" "+prsanitizer.ProwPluginName+" command is triggered by pr creator", func() {
			// given
			NonExistingRawGitHubFiles("pr-sanitizer.yml", "pr-sanitizer.yaml")

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/11").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/pr_details_w_title_type.json"))

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/11/reviews").
				Reply(200).
				BodyString(`[]`)

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/collaborators/bartoszmajsak-test/permission").
				Reply(200).
				BodyString(`{"permission": "read"}`)

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toHaveSuccessState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			event := LoadIssueCommentEvent("test_fixtures/github_calls/trigger_run_pr-sanitizer_on_pr_by_pr_creator.json")

			// when
			err := handler.HandleIssueCommentEvent(log, event)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark status as failed (thus block PR merge) when not prefixed with semantic commit message type when "+command.RunCommentPrefix+" all command is used by admin", func() {
			// given
			NonExistingRawGitHubFiles("pr-sanitizer.yml", "pr-sanitizer.yaml", "work-in-progress.yml", "work-in-progress.yaml")

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/11").
				Reply(200).
				Body(FromFile("test_fixtures/github_calls/pr_details_wo_title_type.json"))

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/pulls/11/reviews").
				Reply(200).
				BodyString(`[]`)

			gock.New("https://api.github.com").
				Get("/repos/" + repositoryName + "/collaborators/bartoszmajsak/permission").
				Reply(200).
				BodyString(`{"permission": "admin"}`)

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/statuses").
				SetMatcher(ExpectPayload(toHaveFailureState)).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			gock.New("https://api.github.com").
				Post("/repos/" + repositoryName + "/issues/11/comments").
				SetMatcher(ExpectPayload(HaveBodyThatContains(fmt.Sprintf("Hey @%s! %s", "bartoszmajsak-test", prsanitizer.TitleFailureMessage)))).
				Reply(201)

			event := LoadIssueCommentEvent("test_fixtures/github_calls/trigger_run_all_on_pr_by_admin.json")

			// when
			err := handler.HandleIssueCommentEvent(log, event)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})
	})
})

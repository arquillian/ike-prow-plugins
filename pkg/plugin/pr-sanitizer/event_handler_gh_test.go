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
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/work-in-progress"
)

const botName = "alien-ike"

var (
	expectedContext = strings.Join([]string{botName, prsanitizer.ProwPluginName}, "/")
	docStatusRoot   = fmt.Sprintf("%s/status/%s", plugin.DocumentationURL, prsanitizer.ProwPluginName)
)

var _ = Describe("PR Sanitizer Plugin features", func() {

	var handler *prsanitizer.GitHubPRSanitizerEventsHandler
	var mocker = NewMockPluginTemplate(prsanitizer.ProwPluginName)

	log := log.NewTestLogger()

	haveSuccessState := SoftlySatisfyAll(
		HaveState(github.StatusSuccess),
		HaveDescription(prsanitizer.SuccessMessage),
		HaveContext(expectedContext),
		HaveTargetURL(fmt.Sprintf("%s/%s/%s.html", docStatusRoot, "success", prsanitizer.SuccessDetailsPageName)),
	)

	haveFailureState := SoftlySatisfyAll(
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
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithTitle("feat: introduces dummy response").
				WithBody("This pr introduces dummy response which is adding new method.\r\n\r\n fixes: #2").
				WithoutConfigFiles().
				Expecting(Status(To(haveSuccessState))).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("opened"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark status as failed (thus block PR merge) when not prefixed with semantic commit message type", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithTitle("introduces dummy response").
				WithBody("This pr introduces dummy response which is adding new method.\r\n\r\n fixes: #2").
				WithoutComments().
				WithoutConfigFiles().
				WithoutConfigFilesForPlugin(wip.ProwPluginName).
				Expecting(Status(To(haveFailureState)), Comment(ContainingStatusMessage(prsanitizer.TitleFailureMessage))).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("opened"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark status as success when title starts with configured semantic commit message type", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithTitle(":star: configures plugin").
				WithBody("This pr introduces dummy response which is adding new method.\r\n\r\n fixes: #2").
				WithConfigFile(ConfigYml(LoadedFrom("test_fixtures/github_calls/pr-sanitizer.yml"))).
				Expecting(Status(To(haveSuccessState))).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("edited"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark status as success (thus unblock PR merge) when title updated to contain semantic commit message type", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithTitle("feat: introduces dummy response").
				WithBody("This pr introduces dummy response which is adding new method.\r\n\r\n fixes: #2").
				WithoutConfigFiles().
				Expecting(Status(To(haveSuccessState))).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("edited"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())

		})

		It("should mark status as success if PR title prefixed with wip and conforms with semantic commit message type", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithTitle("WIP feat: introduces dummy response").
				WithBody("This pr introduces dummy response which is adding new method.\r\n\r\n fixes: #2").
				WithoutConfigFiles().
				WithoutConfigFilesForPlugin(wip.ProwPluginName).
				Expecting(Status(To(haveSuccessState))).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("opened"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark status as failed (thus block PR merge) when prefixed with wip and does not conform with semantic commit message type", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithTitle("WIP introduces dummy response").
				WithBody("This pr introduces dummy response which is adding new method.\r\n\r\n fixes: #2").
				WithoutComments().
				WithoutConfigFiles().
				WithoutConfigFilesForPlugin(wip.ProwPluginName).
				Expecting(Status(To(haveFailureState)), Comment(ContainingStatusMessage(prsanitizer.TitleFailureMessage))).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("opened"))

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
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithTitle("feat: PR from external user without tests should be rejected").
				WithBody("This pr introduces dummy response which is adding new method.\r\n\r\n fixes: #2").
				WithoutReviews().
				WithoutConfigFiles().
				Expecting(Status(To(haveSuccessState))).
				Create()

			// when
			err := handler.HandleIssueCommentEvent(log, prMock.CreateCommentEvent(SentByPrCreator, "/run work-in-progress", "created"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark status as failed (thus block PR merge) when not prefixed with semantic commit message type when "+command.RunCommentPrefix+" all command is used by admin", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithTitle("PR from external user without tests should be rejected").
				WithBody("This pr introduces dummy response which is adding new method.\r\n\r\n fixes: #2").
				WithoutComments().
				WithoutConfigFiles().
				WithoutConfigFilesForPlugin(wip.ProwPluginName).
				WithoutReviews().
				WithUsers(Admin("admin")).
				Expecting(Status(To(haveFailureState)), Comment(ContainingStatusMessage(prsanitizer.TitleFailureMessage))).
				Create()

			// when
			err := handler.HandleIssueCommentEvent(log, prMock.CreateCommentEvent(SentBy("admin"), "/run all", "created"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

	Context("Pull Request Description verifier", func() {

		descriptionContentLengthMessage := fmt.Sprintf(prsanitizer.DescriptionLengthShortMessage, 50)

		BeforeEach(func() {
			defer gock.OffAll()
			handler = &prsanitizer.GitHubPRSanitizerEventsHandler{Client: NewDefaultGitHubClient(), BotName: botName}
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should mark status as failed (thus block PR merge) when PR doesn't have issue linked in the description", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithTitle("feat: introduces dummy response").
				WithBody("This pr introduces dummy response which is adding new method.").
				WithoutComments().
				WithoutConfigFiles().
				WithoutConfigFilesForPlugin(prsanitizer.ProwPluginName).
				Expecting(Status(To(haveFailureState)), Comment(ContainingStatusMessage(prsanitizer.IssueLinkMissingMessage))).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("opened"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark status as failed (thus block PR merge) when PR doesn't have description", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithTitle("introduces dummy response").
				WithBody("fixes: #3").
				WithoutComments().
				WithoutConfigFiles().
				WithoutConfigFilesForPlugin(wip.ProwPluginName).
				Expecting(Status(To(haveFailureState)), Comment(ContainingStatusMessage(fmt.Sprintf("%s\n\n%s", prsanitizer.TitleFailureMessage, descriptionContentLengthMessage)))).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("opened"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark status as failed (thus block PR merge) when PR doesn't have description length as per configuration", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithTitle("feat: introduces dummy response").
				WithBody("This pr introduces dummy response which is adding new method.\r\n\r\n fixes: #3").
				WithConfigFile(
				ConfigYml(Containing(
					Param("description_content_length", "100")))).
				WithoutComments().
				WithoutConfigFiles().
				WithoutConfigFilesForPlugin(wip.ProwPluginName).
				WithoutConfigFilesForPlugin(prsanitizer.ProwPluginName).
				Expecting(Status(To(haveFailureState)), Comment(ContainingStatusMessage(fmt.Sprintf(prsanitizer.DescriptionLengthShortMessage, 100)))).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("opened"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark status as failed (thus block PR merge) when prefixed with wip and does not conform with semantic commit message type and short description", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithTitle("WIP introduces dummy response").
				WithoutComments().
				WithoutConfigFiles().
				WithoutConfigFilesForPlugin(wip.ProwPluginName).
				WithoutConfigFilesForPlugin(prsanitizer.ProwPluginName).
				Expecting(Status(To(haveFailureState)), Comment(ContainingStatusMessage(fmt.Sprintf("%s\n\n%s", prsanitizer.TitleFailureMessage, descriptionContentLengthMessage)))).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("opened"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

	})
})

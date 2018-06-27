package wip_test

import (
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/work-in-progress"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"

	"github.com/arquillian/ike-prow-plugins/pkg/command"
	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper"
)

const botName = "alien-ike"

var _ = Describe("Work In Progress Plugin features", func() {

	var handler *wip.GitHubWIPPRHandler
	var mocker = NewMockPluginTemplate(wip.ProwPluginName)

	log := log.NewTestLogger()

	toHaveSuccessState := ToBe(github.StatusSuccess, wip.ReadyForReviewMessage, wip.ReadyForReviewDetailsPageName)
	toHaveFailureState := ToBe(github.StatusFailure, wip.InProgressMessage, wip.InProgressDetailsPageName)

	Context("Pull Request label change trigger", func() {
		BeforeEach(func() {
			defer gock.OffAll()
			handler = &wip.GitHubWIPPRHandler{Client: NewDefaultGitHubClient(), BotName: botName}
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should mark opened PR as work-in-progress when labeled with WIP", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithTitle("adds a new endpoint.").
				WithoutConfigFiles().
				WithLabels("work-in-progress").
				Expecting(Status(toHaveFailureState)).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("labeled"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark opened PR as ready for review if WIP label removed", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithTitle("[WIP]: feat: adds a new endpoint.").
				WithoutConfigFiles().
				WithoutLabels().
				Expecting(
					Status(toHaveSuccessState),
					ChangedTitle("feat: adds a new endpoint.")).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("unlabeled"))

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
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithTitle("feat: introduces dummy response").
				WithoutConfigFiles().
				WithoutLabels().
				Expecting(
					Status(toHaveSuccessState)).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("opened"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark opened PR as work-in-progress when prefixed with WIP", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithTitle("WIP feat: introduces dummy response").
				WithoutConfigFiles().
				WithoutLabels().
				Expecting(
					AddedLabel("work-in-progress"),
					Status(toHaveFailureState)).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("opened"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark opened PR as work-in-progress when title starts with configured prefix", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithTitle("WORK IN PROGRESS: configures plugin").
				WithConfigFile(ConfigYml(LoadedFrom("test_fixtures/github_calls/work-in-progress.yml"))).
				WithoutLabels().
				Expecting(
					AddedLabel("wip"),
					Status(toHaveFailureState)).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("opened"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark status as failed (thus block PR merge) when title updated to contain WIP", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithTitle("WIP feat: introduces dummy response").
				WithoutConfigFiles().
				WithoutLabels().
				Expecting(
					AddedLabel("work-in-progress"),
					Status(toHaveFailureState)).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("edited"))

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())

		})

		It("should mark status as success (thus unblock PR merge) when title has WIP removed", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithTitle("feat: introduces dummy response").
				WithoutConfigFiles().
				WithLabels("work-in-progress").
				Expecting(
					RemovedLabel("work-in-progress", LoadedFrom("test_fixtures/github_calls/pr_edited_with_unlabel.json")),
					Status(toHaveSuccessState)).
				Create()

			// when
			err := handler.HandlePullRequestEvent(log, prMock.CreatePullRequestEvent("edited"))

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
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithTitle("PR from external user without tests should be rejected").
				WithoutConfigFiles().
				WithoutReviews().
				WithoutLabels().
				WithUsers(ExternalUser("bartoszmajsak")).
				Expecting(
					Status(toHaveSuccessState)).
				Create()

			commentEvent := prMock.CreateCommentEvent(SentByPrCreator, "/run work-in-progress", "created")

			// when
			err := handler.HandleIssueCommentEvent(log, commentEvent)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should mark opened PR as work-in-progress if prefixed with WIP when "+command.RunCommentPrefix+" all command is used by admin", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithTitle("WIP PR from external user without tests should be rejected").
				WithoutConfigFiles().
				WithoutReviews().
				WithoutLabels().
				WithUsers(Admin("bartoszmajsak-test")).
				Expecting(
					AddedLabel("work-in-progress"),
					Status(toHaveFailureState)).
				Create()

			commentEvent := prMock.CreateCommentEvent(SentBy("bartoszmajsak-test"), "/run all", "created")

			// when
			err := handler.HandleIssueCommentEvent(log, commentEvent)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should approve newly created pull request with tests when "+command.RunCommentPrefix+" "+testkeeper.ProwPluginName+" command is triggered by pr creator", func() {
			// given
			prMock := mocker.MockPr().LoadedFromDefaultJSON().
				WithoutConfigFiles().
				WithoutReviews().
				WithUsers(ExternalUser("bartoszmajsak-test")).
				Expecting(NoStatus()).
				Create()

			commentEvent := prMock.CreateCommentEvent(SentByPrCreator, "/run test-keeper", "created")

			// when
			err := handler.HandleIssueCommentEvent(log, commentEvent)

			// then
			Ω(err).ShouldNot(HaveOccurred())
		})
	})

})

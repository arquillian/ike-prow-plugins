package testkeeper

import (
	"github.com/arquillian/ike-prow-plugins/pkg/command"
	ghclient "github.com/arquillian/ike-prow-plugins/pkg/github/client"
	ghservice "github.com/arquillian/ike-prow-plugins/pkg/github/service"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	gogh "github.com/google/go-github/github"
)

// GitHubTestEventsHandler is the event handler for the plugin.
// Implements server.GitHubEventHandler interface which contains the logic for incoming GitHub events
type GitHubTestEventsHandler struct {
	Client  ghclient.Client
	BotName string
}

// ProwPluginName is an external prow plugin name used to register this service
const ProwPluginName = "test-keeper"

var (
	handledPrActions      = []string{"opened", "reopened", "edited", "synchronize"}
	handledCommentActions = []string{"created", "edited", "deleted"}
)

// HandlePullRequestEvent is an entry point for the plugin logic. This method is invoked by the Server when
// pull request event is dispatched from the /hook service
func (gh *GitHubTestEventsHandler) HandlePullRequestEvent(logger log.Logger, event *gogh.PullRequestEvent) error {
	if !utils.Contains(handledPrActions, *event.Action) {
		return nil
	}
	return gh.checkTestsAndSetStatus(logger, ghservice.NewPullRequestLazyLoaderWithPR(gh.Client, event.PullRequest))
}

// HandleIssueCommentEvent is an entry point for the plugin logic. This method is invoked by the Server when
// issue comment event is dispatched from the /hook service
func (gh *GitHubTestEventsHandler) HandleIssueCommentEvent(logger log.Logger, comment *gogh.IssueCommentEvent) error {
	if !utils.Contains(handledCommentActions, *comment.Action) {
		return nil
	}

	prLoader := ghservice.NewPullRequestLazyLoaderFromComment(gh.Client, comment)
	userPerm := command.NewPermissionService(gh.Client, *comment.Sender.Login, prLoader)

	cmdHandler := command.CommentCmdHandler{Client: gh.Client}

	cmdHandler.Register(&command.RunCmd{
		PluginName:            ProwPluginName,
		UserPermissionService: userPerm,
		WhenAddedOrEdited: func() error {
			return gh.checkTestsAndSetStatus(logger, prLoader)
		}})

	cmdHandler.Register(&BypassCmd{
		userPermissionService: userPerm,
		whenDeleted: func() error {
			return gh.checkTestsAndSetStatus(logger, prLoader)
		},
		whenAddedOrEdited: func() error {
			pullRequest, err := prLoader.Load()
			if err != nil {
				return err
			}
			reportBypassCommand(pullRequest)
			statusService := gh.newTestStatusService(logger, pullRequest)
			return statusService.okWithoutTests(*comment.Sender.Login)
		}})

	err := cmdHandler.Handle(logger, comment)
	if err != nil {
		logger.Error(err)
	}
	return err
}

func (gh *GitHubTestEventsHandler) checkIfBypassed(logger log.Logger, commentsLoader *ghservice.IssueCommentsLazyLoader,
	pr *gogh.PullRequest) (found bool, comment string) {
	comments, err := commentsLoader.Load()
	if err != nil {
		logger.Errorf("Getting all comments failed with an error: %s", err)
		return false, ""
	}

	prLoader := ghservice.NewPullRequestLazyLoaderWithPR(gh.Client, pr)
	for _, comment := range comments {
		if IsValidBypassCmd(comment, prLoader) {
			return true, *comment.User.Login
		}
	}
	return false, ""
}

func (gh *GitHubTestEventsHandler) checkTestsAndSetStatus(logger log.Logger, prLoader *ghservice.PullRequestLazyLoader) error {
	pr, err := prLoader.Load()
	if err != nil {
		return err
	}
	change := ghservice.NewRepositoryChangeForPR(pr)
	configuration := LoadConfiguration(logger, change)
	fileCategories, err := gh.checkTests(logger, change, configuration, *pr.Number)
	commentsLoader := ghservice.NewIssueCommentsLazyLoader(gh.Client, pr)

	statusService := gh.newTestStatusServiceWithMessages(logger, pr, commentsLoader, configuration)
	if err != nil {
		if statusErr := statusService.reportError(); statusErr != nil {
			logger.Errorf("failed to report error status on PR [%q]. cause: %s", *pr, statusErr)
		}
		return err
	}

	if fileCategories.OnlySkippedFiles() {
		statusService.onlySkippedMessage()
		return statusService.okOnlySkippedFiles()
	}

	if fileCategories.TestsExist() {
		reportPullRequest(logger, pr, WithTests)
		statusService.withTestsMessage()
		return statusService.okTestsExist()
	}

	bypassed, user := gh.checkIfBypassed(logger, commentsLoader, pr)
	if bypassed {
		reportBypassCommand(pr)
		return statusService.okWithoutTests(user)
	}

	reportPullRequest(logger, pr, WithoutTests)
	statusService.withoutTestsMessage()
	err = statusService.failNoTests()
	if err != nil {
		logger.Errorf("failed to report status on PR [%q]. cause: %s", *pr, err)
	}
	return err
}

func (gh *GitHubTestEventsHandler) checkTests(logger log.Logger, change scm.RepositoryChange,
	config *PluginConfiguration, prNumber int) (FileCategories, error) {
	matcher, err := LoadMatcher(config)
	if err != nil {
		logger.Error(err)
		return FileCategories{}, err
	}

	fileCategoryCounter := FileCategoryCounter{Matcher: matcher}

	changedFiles, err := gh.Client.ListPullRequestFiles(change.Owner, change.RepoName, prNumber)
	if err != nil {
		logger.Error(err)
		return FileCategories{}, err
	}

	fileCategories, err := fileCategoryCounter.Count(changedFiles)
	if err != nil {
		logger.Error(err)
	}

	return fileCategories, err
}

package testkeeper

import (
	"github.com/arquillian/ike-prow-plugins/pkg/command"
	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
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
// events are dispatched from the /hook service
func (gh *GitHubTestEventsHandler) HandlePullRequestEvent(log log.Logger, event *gogh.PullRequestEvent) error {
	if !utils.Contains(handledPrActions, *event.Action) {
		return nil
	}
	return gh.checkTestsAndSetStatus(log, ghservice.NewPullRequestLazyLoaderWithPR(gh.Client, event.PullRequest))
}

// HandleIssueCommentEvent is an entry point for the plugin logic. This method is invoked by the Server when
// events are dispatched from the /hook service
func (gh *GitHubTestEventsHandler) HandleIssueCommentEvent(log log.Logger, comment *gogh.IssueCommentEvent) error {
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
			return gh.checkTestsAndSetStatus(log, prLoader)
		}})

	cmdHandler.Register(&BypassCmd{
		userPermissionService: userPerm,
		whenDeleted: func() error {
			return gh.checkTestsAndSetStatus(log, prLoader)
		},
		whenAddedOrEdited: func() error {
			pullRequest, err := prLoader.Load()
			if err != nil {
				return err
			}
			statusService := gh.newTestStatusService(log, ghservice.NewRepositoryChangeForPR(pullRequest))
			reportBypassCommand(pullRequest)
			return statusService.okWithoutTests(*comment.Sender.Login)
		}})

	err := cmdHandler.Handle(log, comment)
	if err != nil {
		log.Error(err)
	}
	return err
}

func (gh *GitHubTestEventsHandler) checkIfBypassed(log log.Logger, commentsLoader *ghservice.IssueCommentsLazyLoader, pr *gogh.PullRequest) (bool, string) {
	comments, err := commentsLoader.Load()
	if err != nil {
		log.Errorf("Getting all comments failed with an error: %s", err)
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

func (gh *GitHubTestEventsHandler) checkTestsAndSetStatus(log log.Logger, prLoader *ghservice.PullRequestLazyLoader) error {
	pr, err := prLoader.Load()
	if err != nil {
		return err
	}
	change := ghservice.NewRepositoryChangeForPR(pr)
	configuration := LoadConfiguration(log, change)
	fileCategories, err := gh.checkTests(log, change, configuration, *pr.Number)
	statusService := gh.newTestStatusService(log, change)
	if err != nil {
		if statusErr := statusService.reportError(); statusErr != nil {
			log.Errorf("failed to report error status on PR [%q]. cause: %s", *pr, statusErr)
		}
		return err
	}

	if fileCategories.OnlySkippedFiles() {
		return statusService.okOnlySkippedFiles()
	}

	if fileCategories.TestsExist() {
		reportPullRequest(log, pr, WithTests)
		return statusService.okTestsExist()
	}

	commentsLoader := ghservice.NewIssueCommentsLazyLoader(gh.Client, pr)
	bypassed, user := gh.checkIfBypassed(log, commentsLoader, pr)
	if bypassed {
		reportBypassCommand(pr)
		return statusService.okWithoutTests(user)
	}

	reportPullRequest(log, pr, WithoutTests)
	err = statusService.failNoTests()
	if err != nil {
		log.Errorf("failed to report status on PR [%q]. cause: %s", *pr, err)
	}

	hintContext := ghservice.HintContext{PluginName: ProwPluginName, Assignee: *pr.User.Login}
	hinter := ghservice.NewHinter(gh.Client, log, commentsLoader, hintContext)

	cerr := hinter.PluginComment(CreateHintMessage(configuration, change, log))
	if cerr != nil {
		log.Errorf("failed to comment on PR [%q]. cause: %s", *pr, cerr)
		return cerr
	}

	return err
}

func (gh *GitHubTestEventsHandler) checkTests(log log.Logger, change scm.RepositoryChange, config PluginConfiguration, prNumber int) (FileCategories, error) {
	matcher, err := LoadMatcher(config)
	if err != nil {
		log.Error(err)
		return FileCategories{}, err
	}

	fileCategoryCounter := FileCategoryCounter{Matcher: matcher}

	changedFiles, err := gh.Client.ListPullRequestFiles(change.Owner, change.RepoName, prNumber)
	if err != nil {
		log.Error(err)
		return FileCategories{}, err
	}

	fileCategories, err := fileCategoryCounter.Count(changedFiles)
	if err != nil {
		log.Error(err)
	}

	return fileCategories, err
}

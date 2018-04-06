package plugin

import (
	"encoding/json"

	"github.com/arquillian/ike-prow-plugins/pkg/command"
	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	gogh "github.com/google/go-github/github"
)

// GitHubTestEventsHandler is the event handler for the plugin.
// Implements server.GitHubEventHandler interface which contains the logic for incoming GitHub events
type GitHubTestEventsHandler struct {
	Client *github.Client
}

// ProwPluginName is an external prow plugin name used to register this service
const ProwPluginName = "test-keeper"

var (
	handledPrActions      = []string{"opened", "reopened", "edited", "synchronize"}
	handledCommentActions = []string{"created", "deleted"}
)

// HandleEvent is an entry point for the plugin logic. This method is invoked by the Server when
// events are dispatched from the /hook service
func (gh *GitHubTestEventsHandler) HandleEvent(log log.Logger, eventType github.EventType, payload []byte) error {

	switch eventType {
	case github.PullRequest:
		var event gogh.PullRequestEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			log.Errorf("Failed while parsing '%q' event with payload: %q. Cause: %q", github.PullRequest, event, err)
			return err
		}

		if err := gh.handlePrEvent(log, &event); err != nil {
			log.Errorf("Error handling '%q' event with payload %q. Cause: %q", github.PullRequest, event, err)
			return err
		}

	case github.IssueComment:
		var event gogh.IssueCommentEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			log.Errorf("Failed while parsing '%q' event with payload: %q. Cause: %q", github.IssueComment, event, err)
			return err
		}

		if err := gh.handlePrComment(log, &event); err != nil {
			log.Errorf("Error handling '%q' event with payload %q. Cause: %q", github.IssueComment, event, err)
			return err
		}

	default:
		log.Warnf("received an event of type %q but didn't ask for it", eventType)
	}
	return nil
}

func (gh *GitHubTestEventsHandler) handlePrEvent(log log.Logger, event *gogh.PullRequestEvent) error {
	if !utils.Contains(handledPrActions, *event.Action) {
		return nil
	}

	return gh.checkTestsAndSetStatus(log, event.PullRequest)
}

func (gh *GitHubTestEventsHandler) handlePrComment(log log.Logger, prComment *gogh.IssueCommentEvent) error {
	if !utils.Contains(handledCommentActions, *prComment.Action) {
		return nil
	}

	prLoader := github.NewPullRequestLoader(gh.Client, prComment)
	userPerm := command.NewPermissionService(gh.Client, *prComment.Sender.Login, prLoader)

	cmdHandler := command.CommentCmdHandler{Client: gh.Client}

	cmdHandler.Register(
		&SkipCommentCmd{
			userPermissionService: userPerm,
			whenDeleted: func() error {
				pullRequest, err := prLoader.Load()
				if err != nil {
					return err
				}
				return gh.checkTestsAndSetStatus(log, pullRequest)
			},
			whenAddedOrCreated: func() error {
				pullRequest, err := prLoader.Load()
				if err != nil {
					return err
				}
				statusService := gh.newTestStatusService(log, github.NewRepositoryChangeForPR(pullRequest))
				return statusService.okWithoutTests(*prComment.Sender.Login)
			},
		})

	err := cmdHandler.Handle(log, prComment)
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func (gh *GitHubTestEventsHandler) checkTestsAndSetStatus(log log.Logger, pr *gogh.PullRequest) error {
	change := github.NewRepositoryChangeForPR(pr)
	statusService := gh.newTestStatusService(log, change)
	configuration := LoadTestKeeperConfig(log, change)
	fileCategories, err := gh.checkTests(log, change, configuration, *pr.Number)
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
		return statusService.okTestsExist()
	}

	err = statusService.failNoTests()
	if err != nil {
		log.Errorf("failed to report status on PR [%q]. cause: %s", *pr, err)
	}

	hintContext := github.HintContext{PluginName: ProwPluginName, Assignee: *pr.User.Login}
	hinter := github.NewHinter(gh.Client, log, change, *pr.Number, hintContext)

	cerr := hinter.PluginComment(CreateCommentMessage(configuration, change))
	if cerr != nil {
		log.Errorf("failed to comment on PR [%q]. cause: %s", *pr, cerr)
		return cerr
	}

	return err
}

func (gh *GitHubTestEventsHandler) checkTests(log log.Logger, change scm.RepositoryChange, config TestKeeperConfiguration, prNumber int) (FileCategories, error) {
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

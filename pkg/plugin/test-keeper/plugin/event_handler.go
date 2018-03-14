package plugin

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	gogh "github.com/google/go-github/github"
)

// GitHubTestEventsHandler is the event handler for the plugin.
// Implements server.GitHubEventHandler interface which contains the logic for incoming GitHub events
type GitHubTestEventsHandler struct {
	Client *gogh.Client
}

// ProwPluginName is an external prow plugin name used to register this service
const (
	ProwPluginName = "test-keeper"
	SkipComment    = "/ok-without-tests"
)

var (
	handledPrActions      = []string{"opened", "reopened", "synchronize"}
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

	change := scm.RepositoryChange{
		Owner:    *event.Repo.Owner.Login,
		RepoName: *event.Repo.Name,
		Hash:     *event.PullRequest.Head.SHA,
	}

	return gh.checkTestsAndSetStatus(log, change, event.PullRequest)
}

func (gh *GitHubTestEventsHandler) handlePrComment(log log.Logger, prComment *gogh.IssueCommentEvent) error {
	if !utils.Contains(handledCommentActions, *prComment.Action) {
		return nil
	}

	org := prComment.Repo.Owner.Login
	name := prComment.Repo.Name
	prNumber := prComment.Issue.Number

	sender := prComment.Sender.Login
	permissionLevel, _, e := gh.Client.Repositories.GetPermissionLevel(context.Background(), *org, *name, *sender)
	if e != nil {
		log.Fatal(e)
		return e
	}

	if *permissionLevel.Permission != "admin" {
		return nil
	}

	comment := strings.TrimSpace(*prComment.Comment.Body)

	if comment != SkipComment {
		return nil
	}

	pullRequest, _, err := gh.Client.PullRequests.Get(context.Background(), *org, *name, *prNumber)
	if err != nil {
		log.Fatal(err)
		return err
	}

	change := scm.RepositoryChange{
		Owner:    *prComment.Repo.Owner.Login,
		RepoName: *prComment.Repo.Name,
		Hash:     *pullRequest.Head.SHA,
	}

	if comment == SkipComment && *prComment.Action == "deleted" {
		return gh.checkTestsAndSetStatus(log, change, pullRequest)
	}

	statusService := gh.newTestStatusService(log, change)
	return statusService.okWithoutTests(*sender)
}

func (gh *GitHubTestEventsHandler) checkTestsAndSetStatus(log log.Logger, change scm.RepositoryChange, pr *gogh.PullRequest) error {
	configuration := LoadTestKeeperConfig(log, change)
	testsExist, err := gh.checkTests(log, change, configuration, *pr.Number)
	if err != nil {
		return err
	}

	statusService := gh.newTestStatusService(log, change)
	if testsExist {
		return statusService.testsExist()
	}

	err = statusService.noTests()
	if err != nil {
		log.Error("There occur an error when the status was being set to PR:", err)
	}

	commentContext := plugin.CommentContext{PluginName: ProwPluginName, Assignee: *pr.User.Login}
	commentService := plugin.NewCommentService(gh.Client, log, change, *pr.Number, commentContext)

	cerr := commentService.PluginComment(CreateCommentMessage(configuration, change))
	if cerr != nil {
		return cerr
	}
	return err
}

func (gh *GitHubTestEventsHandler) checkTests(log log.Logger, change scm.RepositoryChange, config TestKeeperConfiguration, prNumber int) (bool, error) {
	matcher := LoadMatcher(config)

	checker := TestChecker{TestKeeperMatcher: matcher}

	files, _, err := gh.Client.PullRequests.ListFiles(context.Background(), change.Owner, change.RepoName, prNumber, nil)
	if err != nil {
		log.Error(err)
		return false, nil
	}

	testsExist, err := checker.IsAnyNotExcludedFileTest(asChangedFiles(files))
	if err != nil {
		log.Error(err)
	}

	return testsExist, err
}

func asChangedFiles(files []*gogh.CommitFile) []scm.ChangedFile {
	changedFiles := make([]scm.ChangedFile, 0, len(files))
	for _, file := range files {
		changedFiles = append(changedFiles, scm.ChangedFile{Name: *file.Filename, Status: *file.Status})
	}

	return changedFiles
}

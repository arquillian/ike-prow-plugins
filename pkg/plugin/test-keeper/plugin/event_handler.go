package plugin

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	gogh "github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
)

// GitHubTestEventsHandler is the event handler for the plugin.
// Implements server.GitHubEventHandler interface which contains the logic for incoming GitHub events
type GitHubTestEventsHandler struct {
	Client *gogh.Client
	Log    *logrus.Entry
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
func (gh *GitHubTestEventsHandler) HandleEvent(eventType github.EventType, eventGUID string, payload []byte) error {
	if gh.Log == nil {
		gh.Log = logrus.StandardLogger().WithField("ike-plugins", ProwPluginName).WithFields(
			logrus.Fields{
				"event-type": eventType,
				"event-GUID": eventGUID,
			},
		)
	}
	switch eventType {
	case github.PullRequest:
		gh.Log.Info("Pull request received")
		var event gogh.PullRequestEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			gh.Log.Errorf("Failed while parsing '%q' event with payload: %q. Cause: %q", github.PullRequest, event, err)
			return err
		}

		if err := gh.handlePrEvent(&event); err != nil {
			gh.Log.Errorf("Error handling '%q' event with payload %q. Cause: %q", github.PullRequest, event, err)
			return err
		}

	case github.IssueComment:
		gh.Log.Info("Issue comment event.")
		var event gogh.IssueCommentEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			gh.Log.Errorf("Failed while parsing '%q' event with payload: %q. Cause: %q", github.IssueComment, event, err)
			return err
		}

		if err := gh.handlePrComment(&event); err != nil {
			gh.Log.Errorf("Error handling '%q' event with payload %q. Cause: %q", github.IssueComment, event, err)
			return err
		}

	default:
		gh.Log.Infof("received an event of type %q but didn't ask for it", eventType)
	}
	return nil
}

func (gh *GitHubTestEventsHandler) handlePrEvent(event *gogh.PullRequestEvent) error {
	if !utils.Contains(handledPrActions, *event.Action) {
		gh.Log.Info("No proper action", *event.Action)
		return nil
	}

	change := scm.RepositoryChange{
		Owner:    *event.Repo.Owner.Login,
		RepoName: *event.Repo.Name,
		Hash:     *event.PullRequest.Head.SHA,
	}

	return gh.checkTestsAndSetStatus(change, event.PullRequest)
}

func (gh *GitHubTestEventsHandler) handlePrComment(prComment *gogh.IssueCommentEvent) error {
	if !utils.Contains(handledCommentActions, *prComment.Action) {
		gh.Log.Info("No proper action", *prComment.Action)
		return nil
	}

	org := prComment.Repo.Owner.Login
	name := prComment.Repo.Name
	prNumber := prComment.Issue.Number
	gh.Log.Infof("Received event, %q, %q, %d", *org, *name, *prNumber)

	sender := prComment.Sender.Login
	permissionLevel, _, e := gh.Client.Repositories.GetPermissionLevel(context.Background(), *org, *name, *sender)
	if e != nil {
		gh.Log.Fatal(e)
		return e
	}

	if *permissionLevel.Permission != "admin" {
		gh.Log.Infof("%q does not have admin permission to accept PR without tests", *sender)
		return nil
	}

	comment := strings.TrimSpace(*prComment.Comment.Body)

	if comment != SkipComment {
		gh.Log.Infof("'%q' is not a supported comment", comment)
		return nil
	}

	pullRequest, _, err := gh.Client.PullRequests.Get(context.Background(), *org, *name, *prNumber)
	if err != nil {
		gh.Log.Fatal(err)
		return err
	}

	change := scm.RepositoryChange{
		Owner:    *prComment.Repo.Owner.Login,
		RepoName: *prComment.Repo.Name,
		Hash:     *pullRequest.Head.SHA,
	}

	if comment == SkipComment && *prComment.Action == "deleted" {
		return gh.checkTestsAndSetStatus(change, pullRequest)
	}

	statusService := gh.newTestStatusService(change)
	return statusService.okWithoutTests(*sender)
}

func (gh *GitHubTestEventsHandler) checkTestsAndSetStatus(change scm.RepositoryChange, pr *gogh.PullRequest) error {
	configuration := LoadTestKeeperConfig(gh.Log, change)
	testsExist, err := gh.checkTests(change, configuration, *pr.Number)
	if err != nil {
		return err
	}

	statusService := gh.newTestStatusService(change)
	if testsExist {
		return statusService.testsExist()
	}

	err = statusService.noTests()
	if err != nil {
		gh.Log.Error("There occur an error when the status was being set to PR:", err)
	}

	commentContext := plugin.CommentContext{PluginName: ProwPluginName, Assignee: *pr.User.Login}
	commentService := plugin.NewCommentService(gh.Client, gh.Log, change, *pr.Number, commentContext)

	cerr := commentService.PluginComment(CreateCommentMessage(configuration, change))
	if cerr != nil {
		return cerr
	}
	return err
}

func (gh *GitHubTestEventsHandler) checkTests(change scm.RepositoryChange, config TestKeeperConfiguration, prNumber int) (bool, error) {
	matcher := LoadMatcher(config)

	checker := TestChecker{Log: gh.Log, TestKeeperMatcher: matcher}

	files, _, err := gh.Client.PullRequests.ListFiles(context.Background(), change.Owner, change.RepoName, prNumber, nil)
	if err != nil {
		gh.Log.Error(err)
		return false, nil
	}

	testsExist, err := checker.IsAnyNotExcludedFileTest(asChangedFiles(files))
	if err != nil {
		gh.Log.Error(err)
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

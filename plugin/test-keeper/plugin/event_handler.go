package plugin

import (
	"github.com/sirupsen/logrus"
	"fmt"
	"github.com/arquillian/ike-prow-plugins/plugin/utils"
	gogh "github.com/google/go-github/github"
	"encoding/json"
	"context"
	"strings"
	"github.com/arquillian/ike-prow-plugins/plugin/github"
	"github.com/arquillian/ike-prow-plugins/plugin/scm"
	"github.com/arquillian/ike-prow-plugins/plugin/config"
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
	handledPrActions      = []string{"opened", "closed", "reopened", "synchronize", "edited"}
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

	ok, e := gh.checkTests(change, *event.PullRequest.Number)
	if e != nil {
		return e
	}

	statusContext := github.StatusContext{BotName: "ike-plugins", PluginName: ProwPluginName}
	statusService := github.NewStatusService(gh.Client, gh.Log, change, statusContext)
	var err error
	if ok {
		err = statusService.Success("There are some tests :)")
	} else {
		err = statusService.Failure("No tests in this PR :(")
	}

	return err
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
	statusContext := github.StatusContext{BotName: "ike-plugins", PluginName: ProwPluginName}
	statusService := github.NewStatusService(gh.Client, gh.Log, change, statusContext)

	if comment == SkipComment && *prComment.Action == "deleted" {
		ok, _ := gh.checkTests(change, *prNumber)
		if ok {
			statusService.Success("There are some tests :)")
		} else {
			statusService.Failure("No tests in this PR :(")
		}
	}

	// TODO add comment mentioning lack of tests
	statusService.Success(fmt.Sprintf("PR is fine without tests says @%s", *sender))

	return nil
}

func (gh *GitHubTestEventsHandler) checkTests(change scm.RepositoryChange, prNumber int) (bool, error) {
	configLoader := config.PluginConfigLoader{
		PluginName: ProwPluginName,
		Change:     change,
	}

	configuration := TestKeeperConfiguration{}
	configLoader.Load(&configuration)

	matchers := LoadMatchers(configuration, func() []string {
		repositoryService := &github.RepositoryService{
			Client: gh.Client,
			Change: change,
		}
		languages, e := repositoryService.UsedLanguages()
		if e != nil {
			gh.Log.Warnf("No languages found for %s", change)
			return []string{}
		}

		return languages
	})

	checker := TestChecker{
		log: gh.Log,
		testMatchers: matchers,
	}

	files, _, err := gh.Client.PullRequests.ListFiles(context.Background(), change.Owner, change.RepoName, prNumber, nil)

	testsExist, err := checker.IsAnyTestPresent(asChangedFiles(files))
	if err != nil {
		gh.Log.Error(err)
		return false, err
	}

	return testsExist, nil
}
func asChangedFiles(files []*gogh.CommitFile) []scm.ChangedFile {
	changedFiles := make([]scm.ChangedFile, len(files))
	for _, file := range files {
		changedFiles = append(changedFiles, scm.ChangedFile{Name: *file.Filename, Status: *file.Status})
	}

	return changedFiles
}



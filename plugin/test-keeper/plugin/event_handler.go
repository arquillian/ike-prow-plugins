package plugin

import (
	"github.com/sirupsen/logrus"
	"fmt"
	"github.com/arquillian/ike-prow-plugins/plugin/utils"
	"github.com/google/go-github/github"
	"encoding/json"
	"context"
	"strings"
	"github.com/arquillian/ike-prow-plugins/plugin/github"
	"github.com/arquillian/ike-prow-plugins/scm"
)

// GitHubTestEventsHandler is the event handler for the plugin.
// Implements server.GitHubEventHandler interface which contains the logic for incoming GitHub events
type GitHubTestEventsHandler struct {
	Client *github.Client
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
func (gh *GitHubTestEventsHandler) HandleEvent(eventType githubevents.EventType, eventGUID string, payload []byte) error {
	if gh.Log == nil {
		gh.Log = logrus.StandardLogger().WithField("ike-plugins", ProwPluginName).WithFields(
			logrus.Fields{
				"event-type": eventType,
				"event-GUID": eventGUID,
			},
		)
	}
	switch eventType {
	case githubevents.PullRequest:
		gh.Log.Info("Pull request received")
		var event github.PullRequestEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			gh.Log.Errorf("Failed while parsing '%q' event with payload: %q. Cause: %q", githubevents.PullRequest, event, err)
			return err
		}

		if err := gh.handlePrEvent(&event); err != nil {
			gh.Log.Errorf("Error handling '%q' event with payload %q. Cause: %q", githubevents.PullRequest, event, err)
			return err
		}

	case githubevents.IssueComment:
		gh.Log.Info("Issue comment event.")
		var event github.IssueCommentEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			gh.Log.Errorf("Failed while parsing '%q' event with payload: %q. Cause: %q", githubevents.IssueComment, event, err)
			return err
		}

		if err := gh.handlePrComment(&event); err != nil {
			gh.Log.Errorf("Error handling '%q' event with payload %q. Cause: %q", githubevents.IssueComment, event, err)
			return err
		}

	default:
		gh.Log.Infof("received an event of type %q but didn't ask for it", eventType)
	}
	return nil
}

func (gh *GitHubTestEventsHandler) handlePrEvent(prEvent *github.PullRequestEvent) error {
	if !utils.Contains(handledPrActions, *prEvent.Action) {
		gh.Log.Info("No proper action", *prEvent.Action)
		return nil
	}

	return gh.checkTests(gh.createGitPR(prEvent.Repo, prEvent.PullRequest))
}

func (gh *GitHubTestEventsHandler) checkTests(pr scm.CommitScmService) error {
	checker := TestChecker{
		Log:           gh.Log,
		CommitService: pr,
	}
	ok, e := checker.IsAnyTestPresent()
	if e != nil {
		gh.Log.Fatal(e)
		return e
	}
	if ok {
		pr.Success("There are some tests :)")
	} else {
		pr.Fail("No tests in this PR :(")
	}
	return nil
}

func (gh *GitHubTestEventsHandler) handlePrComment(prComment *github.IssueCommentEvent) error {
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

	pr := gh.createGitPR(prComment.Repo, pullRequest)

	if comment == SkipComment && *prComment.Action == "deleted" {
		return gh.checkTests(pr)
	}

	// TODO add comment mentioning lack of tests
	pr.Success(fmt.Sprintf("PR is fine without tests says @%s", *sender))

	return nil
}

func (gh *GitHubTestEventsHandler) createGitPR(repo *github.Repository, pullRequest *github.PullRequest) scm.CommitScmService {
	return scm.CreateScmCommitService(gh.Client, gh.Log, repo, *pullRequest.Head.SHA)
}

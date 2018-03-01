package plugin

import (
	"github.com/sirupsen/logrus"
	"regexp"
	"fmt"
	"github.com/arquillian/ike-prow-plugins/plugin/utils"
	"github.com/google/go-github/github"
	"encoding/json"
	"context"
	"strings"
	"github.com/arquillian/ike-prow-plugins/plugin/github"
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
			gh.Log.Info("JSON: ", event)
			return err
		}

		if err := gh.handlePrEvent(&event); err != nil {
			gh.Log.Error("Error handling event.")
		}

	case githubevents.IssueComment:
		gh.Log.Info("Issue comment event.")
		var event github.IssueCommentEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			return err
		}

		if err := gh.handlePrComment(&event); err != nil {
			gh.Log.Error("Error handling event.")
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

	return gh.checkTests(prEvent.Repo.Owner.Login, prEvent.Repo.Name, prEvent.PullRequest.Head.SHA, prEvent.Number)
}

// TODO refactor to its own testable component
func (gh *GitHubTestEventsHandler) checkTests(org, name, sha *string, prNumber *int) error {

	files, _, e := gh.Client.PullRequests.ListFiles(context.Background(), *org, *name, *prNumber, nil)
	if e != nil {
		gh.Log.Fatal(e)
		return e
	}

	var status = "failure"
	var reason = "No tests in this PR :("
	for _, file := range files {
		// TODO status must be added or changed
		if regexp.MustCompile(`.+(IT\.java|Test.java)$`).MatchString(*file.Filename) {
			status = "success"
			reason = "There are some tests :)"
		}
	}

	if _, _, err := gh.Client.Repositories.CreateStatus(context.Background(), *org, *name, *sha, &github.RepoStatus{
		State:       &status,
		Context:     utils.String("alien-ike/prow-spike"),
		Description: &reason,
	}); err != nil {
		gh.Log.Info("Error handling event.", err)
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
	sha := pullRequest.Head.SHA

	if comment == SkipComment && *prComment.Action == "deleted" {
		return gh.checkTests(org, name, sha, prNumber)
	}

	// TODO add comment mentioning lack of tests

	if _, _, err := gh.Client.Repositories.CreateStatus(context.Background(), *org, *name, *sha, &github.RepoStatus{
		State:       utils.String("success"),
		Context:     utils.String("alien-ike/prow-spike"),
		Description: utils.String(fmt.Sprintf("PR is fine without tests says @%s", *sender)),
	}); err != nil {
		gh.Log.Info("Error handling event.", err)
	}

	return nil
}

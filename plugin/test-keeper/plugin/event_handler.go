package plugin

import (
	"github.com/sirupsen/logrus"
	"regexp"
	"fmt"
	. "github.com/arquillian/ike-prow-plugins/plugin/utils"
	"github.com/google/go-github/github"
	"encoding/json"
	"context"
	"strings"
)

// GitHubTestEventsHandler is the event handler for the plugin.
// Implements server.GitHubEventHandler interface which contains the logic for incoming GitHub events
type GitHubTestEventsHandler struct {
	Client *github.Client
	log *logrus.Entry
}

const ProwPluginName = "test-keeper"

var (
	handledPrActions = []string{"opened", "closed", "reopened", "synchronize", "edited"}
	handledCommentActions = []string{"created", "deleted"}
)

// HandleEvent is an entry point for the plugin logic. This method is invoked by the Server when
// events are dispatched from the /hook service
func (gh *GitHubTestEventsHandler) HandleEvent(eventType, eventGUID string, payload []byte) error {
	gh.log = logrus.StandardLogger().WithField("ike-plugins", ProwPluginName).WithFields(
		logrus.Fields{
			"event-type": eventType,
			"event-GUID": eventGUID,
		},
	)
	switch eventType {
		case "pull_request":
			gh.log .Info("Pull request received")
			var event github.PullRequestEvent
			if err := json.Unmarshal(payload, &event); err != nil {
				return err
			}

			go func() {
				if err := gh.handlePrEvent(&event); err != nil {
					gh.log.Error("Error handling event.")
				}
			}()

		case "issue_comment":
			gh.log.Info("Handling issue comment event.")
			var event github.IssueCommentEvent
			if err := json.Unmarshal(payload, &event); err != nil {
				return err
			}

			go func() {
				if err := gh.handlePrComment(&event); err != nil {
					gh.log.Error("Error handling event.")
				}
			}()
		default:
			gh.log.Infof("received an event of type %q but didn't ask for it", eventType)
	}
	return nil
}

func (gh *GitHubTestEventsHandler) handlePrEvent(prEvent *github.PullRequestEvent) error {
	gh.log.Infof("PR Event %q", *prEvent.Action)
	if !Contains(handledPrActions, *prEvent.Action) {
		return nil
	}

	return gh.checkTests(prEvent.Repo.Owner.Login, prEvent.Repo.Name, prEvent.PullRequest.Head.SHA, prEvent.Number)
}

// TODO refactor to its own testable component
func (gh *GitHubTestEventsHandler) checkTests(org, name, sha *string, prNumber *int) error {

	gh.log.Infof("Checking for tests")
	files, _, e := gh.Client.PullRequests.ListFiles(context.Background(), *org, *name, *prNumber, nil); if e != nil {
		gh.log.Fatal(e)
		return e
	}

	var status = "failure"
	var reason = "No tests in this PR :("
	for _, file := range files {
		gh.log.Infof("%q: %q", *file.Filename, *file.Status)
		// TODO status must be added or changed
		if regexp.MustCompile(`.+(IT\.java|Test.java)$`).MatchString(*file.Filename) == true {
			status = "success"
			reason = "There are some tests :)"
		}
	}

	if _, _, err := gh.Client.Repositories.CreateStatus(context.Background(), *org, *name, *sha, &github.RepoStatus{
		State: &status,
		Context: String("alien-ike/prow-spike"),
		Description: &reason,
	}); err != nil {
		gh.log.Info("Error handling event.", err)
	}

	return nil
}

func (gh *GitHubTestEventsHandler) handlePrComment(prComment *github.IssueCommentEvent) error {
	if !Contains(handledCommentActions, *prComment.Action) {
		return nil
	}

	org := prComment.Repo.Owner.Login
	name := prComment.Repo.Name
	prNumber := prComment.Issue.Number
	gh.log.Infof("Received event, %q, %q, %q", *org, *name, *prNumber)

	sender := prComment.Sender.Login
	permissionLevel, _, e := gh.Client.Repositories.GetPermissionLevel(context.Background(), *org, *name, *sender)
	if e != nil {
		gh.log.Fatal(e)
		return e
	}

	if *permissionLevel.Permission != "admin" {
		gh.log.Infof("%q does not have admin permission to accept PR without tests", *sender)
		return nil
	}

	comment := strings.TrimSpace(*prComment.Comment.Body)

	if comment != "/ok-without-tests"  {
		gh.log.Infof("'%q' is not a supported comment", comment)
		return nil
	}

	pullRequest, _, err := gh.Client.PullRequests.Get(context.Background(), *org, *name, *prNumber)
	if err != nil {
		gh.log.Fatal(err)
		return err
	}
	sha := pullRequest.Head.SHA

	if comment == "/ok-without-tests" && *prComment.Action == "deleted" {
		return gh.checkTests(org, name, sha, prNumber)
	}

	// TODO add comment mentioning lack of tests

	if _, _, err := gh.Client.Repositories.CreateStatus(context.Background(), *org, *name, *sha, &github.RepoStatus{
		State:       String("success"),
		Context:     String("alien-ike/prow-spike"),
		Description: String(fmt.Sprintf("PR is fine without tests says @%s", *sender)),
	}); err != nil {
		gh.log.Info("Error handling event.", err)
	}

	return nil
}

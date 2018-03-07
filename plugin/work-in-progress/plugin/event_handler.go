package plugin

import (
	"github.com/sirupsen/logrus"
	"github.com/google/go-github/github"
	"strings"
	"github.com/arquillian/ike-prow-plugins/plugin/github"
	"encoding/json"
	"github.com/arquillian/ike-prow-plugins/plugin/utils"
	"github.com/arquillian/ike-prow-plugins/scm"
)

// ProwPluginName is an external prow plugin name used to register this service
const ProwPluginName = "work-in-progress"

// GitHubWIPPRHandler handles PR events and updates status of the PR based on work-in-progress indicator
type GitHubWIPPRHandler struct {
	Client *github.Client
	Log    *logrus.Entry
}

var (
	handledPrActions = []string{"opened", "reopened", "edited"}
)

// HandleEvent is an entry point for the plugin logic. This method is invoked by the Server when
// events are dispatched from the /hook service
func (gh *GitHubWIPPRHandler) HandleEvent(eventType githubevents.EventType, eventGUID string, payload []byte) error {
	if gh.Log == nil {
		gh.Log = logrus.StandardLogger().WithField("ike-plugins", ProwPluginName).WithFields(
			logrus.Fields{
				"event-type": eventType,
				"event-GUID": eventGUID,
			},
		)
	}

	gh.Log.Info("Handling title change event.")

	switch eventType {
	case githubevents.PullRequest:
		gh.Log.Info("Pull request received")
		var event github.PullRequestEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			gh.Log.Errorf("Failed while parsing '%q' event with payload: %q. Cause: %q", githubevents.PullRequest, event, err)
			return err
		}

		if !utils.Contains(handledPrActions, *event.Action) {
			gh.Log.Info("No proper action", *event.Action)
			return nil
		}

		head := scm.CreateScmCommitService(gh.Client, gh.Log, event.Repo, *event.PullRequest.Head.SHA)
		if gh.IsWorkInProgress(event.PullRequest.Title) {
			head.Fail("PR is in progress and can't be merged yet. You might want to wait with review as well")
		} else {
			head.Success("PR is ready for review and merge")
		}

	default:
		gh.Log.Infof("received an event of type %q but didn't ask for it", eventType)
	}

	return nil
}

// IsWorkInProgress checks if title is marked as Work In Progress
func (gh *GitHubWIPPRHandler) IsWorkInProgress(title *string) bool {
	return strings.HasPrefix(strings.ToLower(*title), "wip ")
}

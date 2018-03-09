package plugin

import (
	"github.com/sirupsen/logrus"
	gogh "github.com/google/go-github/github"
	"strings"
	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"encoding/json"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
)

// ProwPluginName is an external prow plugin name used to register this service
const ProwPluginName = "work-in-progress"

// GitHubWIPPRHandler handles PR events and updates status of the PR based on work-in-progress indicator
type GitHubWIPPRHandler struct {
	Client *gogh.Client
	Log    *logrus.Entry
}

var (
	handledPrActions = []string{"opened", "reopened", "edited"}
)

// HandleEvent is an entry point for the plugin logic. This method is invoked by the Server when
// events are dispatched from the /hook service
func (gh *GitHubWIPPRHandler) HandleEvent(eventType github.EventType, eventGUID string, payload []byte) error {
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
	case github.PullRequest:
		gh.Log.Info("Pull request received")
		var event gogh.PullRequestEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			gh.Log.Errorf("Failed while parsing '%q' event with payload: %q. Cause: %q", github.PullRequest, event, err)
			return err
		}

		if !utils.Contains(handledPrActions, *event.Action) {
			gh.Log.Info("No proper action", *event.Action)
			return nil
		}

		change := scm.RepositoryChange{
			Owner:    *event.Repo.Owner.Login,
			RepoName: *event.Repo.Name,
			Hash:     *event.PullRequest.Head.SHA,
		}
		statusContext := github.StatusContext{BotName: "ike-plugins", PluginName: ProwPluginName}
		statusService := github.NewStatusService(gh.Client, gh.Log, change, statusContext)
		if gh.IsWorkInProgress(event.PullRequest.Title) {
			return statusService.Failure("PR is in progress and can't be merged yet. You might want to wait with review as well")
		}
		return statusService.Success("PR is ready for review and merge")

	default:
		gh.Log.Infof("received an event of type %q but didn't ask for it", eventType)
	}

	return nil
}

// IsWorkInProgress checks if title is marked as Work In Progress
func (gh *GitHubWIPPRHandler) IsWorkInProgress(title *string) bool {
	return strings.HasPrefix(strings.ToLower(*title), "wip ")
}

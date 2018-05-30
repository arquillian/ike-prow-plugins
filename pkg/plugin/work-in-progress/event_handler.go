package wip

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	gogh "github.com/google/go-github/github"
)

const (
	// ProwPluginName is an external prow plugin name used to register this service
	ProwPluginName = "work-in-progress"

	// InProgressMessage is a message used in GH Status as description when the PR is in progress
	InProgressMessage = "PR is in progress and can't be merged yet. You might want to wait with review as well"
	// InProgressDetailsPageName is a name of a documentation page that contains additional status details for InProgressMessage
	InProgressDetailsPageName = "wip-failed"

	// ReadyForReviewMessage is a message used in GH Status as description when the PR is ready for review and merge
	ReadyForReviewMessage = "PR is ready for review and merge"
	// ReadyForReviewDetailsPageName is a name of a documentation page that contains additional status details for ReadyForReviewMessage
	ReadyForReviewDetailsPageName = "wip-success"
)

// GitHubWIPPRHandler handles PR events and updates status of the PR based on work-in-progress indicator
type GitHubWIPPRHandler struct {
	Client  ghclient.Client
	BotName string
}

var (
	handledPrActions = []string{"opened", "reopened", "edited", "synchronize"}
	defaultPrefixes  = []string{"WIP", "DO NOT MERGE", "DON'T MERGE", "WORK-IN-PROGRESS"}
)

// HandleEvent is an entry point for the plugin logic. This method is invoked by the Server when
// events are dispatched from the /hook service
func (gh *GitHubWIPPRHandler) HandleEvent(log log.Logger, eventType github.EventType, payload []byte) error {
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

	default:
		log.Warnf("received an event of type %q but didn't ask for it", eventType)
	}

	return nil
}

func (gh *GitHubWIPPRHandler) handlePrEvent(log log.Logger, event *gogh.PullRequestEvent) error {
	if !utils.Contains(handledPrActions, *event.Action) {
		return nil
	}

	change := scm.RepositoryChange{
		Owner:    *event.Repo.Owner.Login,
		RepoName: *event.Repo.Name,
		Hash:     *event.PullRequest.Head.SHA,
	}
	statusContext := github.StatusContext{BotName: gh.BotName, PluginName: ProwPluginName}
	statusService := ghservice.NewStatusService(gh.Client, log, change, statusContext)

	labels, err := gh.Client.ListPullRequestLabels(change, *event.PullRequest.Number)
	if err != nil {
		log.Warnf("failed to list labels on PR [%q]. cause: %s", *event.PullRequest, err)
	}

	configuration := LoadConfiguration(log, change)
	labelExists := gh.hasWorkInProgressLabel(labels, configuration.Label)
	if gh.IsWorkInProgress(*event.PullRequest.Title, configuration) {
		if !labelExists {
			if err := gh.Client.AddPullRequestLabel(change, *event.PullRequest.Number, strings.Fields(configuration.Label)); err != nil {
				log.Errorf("failed to add label on PR [%q]. cause: %s", *event.PullRequest, err)
			}
		}
		return statusService.Failure(InProgressMessage, InProgressDetailsPageName)
	}
	if labelExists {
		if err := gh.Client.RemovePullRequestLabel(change, *event.PullRequest.Number, configuration.Label); err != nil {
			log.Errorf("failed to remove label on PR [%q]. cause: %s", *event.PullRequest, err)
		}
	}
	return statusService.Success(ReadyForReviewMessage, ReadyForReviewDetailsPageName)
}

func (gh *GitHubWIPPRHandler) hasWorkInProgressLabel(labels []*gogh.Label, wipLabel string) bool {
	for _, label := range labels {
		if label.GetName() == wipLabel {
			return true
		}
	}
	return false
}

// IsWorkInProgress checks if title is marked as Work In Progress
func (gh *GitHubWIPPRHandler) IsWorkInProgress(title string, config PluginConfiguration) bool {
	prefixes := defaultPrefixes
	if len(config.Prefix) != 0 {
		if config.Combine {
			prefixes = append(prefixes, config.Prefix...)
		} else {
			prefixes = config.Prefix
		}
	}
	return gh.hasPrefix(strings.ToLower(title), prefixes)
}

func (gh *GitHubWIPPRHandler) hasPrefix(title string, prefixes []string) bool {
	for _, prefix := range prefixes {
		pattern := `(?mi)^(\[|\()?` + prefix + `(\]|\))?(:| )+`
		if match, _ := regexp.MatchString(pattern, title); match {
			return true
		}
	}
	return false
}

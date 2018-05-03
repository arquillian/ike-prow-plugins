package wip

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	gogh "github.com/google/go-github/github"
)

const (
	// ProwPluginName is an external prow plugin name used to register this service
	ProwPluginName = "work-in-progress"

	// InProgressMessage is a message used in GH Status as description when the PR is in progress
	InProgressMessage = "PR is in progress and can't be merged yet. You might want to wait with review as well"
	// InProgressDetailsLink is a link to an anchor in arq documentation that contains additional status details for InProgressMessage
	InProgressDetailsLink = plugin.DocumentationURL + "#wip-failed"

	// ReadyForReviewMessage is a message used in GH Status as description when the PR is ready for review and merge
	ReadyForReviewMessage = "PR is ready for review and merge"
	// ReadyForReviewDetailsLink is a link to an anchor in arq documentation that contains additional status details for ReadyForReviewMessage
	ReadyForReviewDetailsLink = plugin.DocumentationURL + "#wip-success"
)

// GitHubWIPPRHandler handles PR events and updates status of the PR based on work-in-progress indicator
type GitHubWIPPRHandler struct {
	Client  ghclient.Client
	BotName string
}

var (
	handledPrActions = []string{"opened", "reopened", "edited"}
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
		configuration := LoadConfiguration(log, change)
		if gh.IsWorkInProgress(event.PullRequest.Title, configuration) {
			return statusService.Failure(InProgressMessage, InProgressDetailsLink)
		}
		return statusService.Success(ReadyForReviewMessage, ReadyForReviewDetailsLink)

	default:
		log.Warnf("received an event of type %q but didn't ask for it", eventType)
	}

	return nil
}

// IsWorkInProgress checks if title is marked as Work In Progress
func (gh *GitHubWIPPRHandler) IsWorkInProgress(title *string, config PluginConfiguration) bool {
	transformedTitle := strings.ToLower(*title)
	if (len(config.Prefix)) > 0 {
		defaultPrefixes = append(defaultPrefixes, config.Prefix...)
	}
	return gh.hasDefaultPrefix(transformedTitle)
}

func (gh *GitHubWIPPRHandler) hasDefaultPrefix(title string) bool {
	for _, prefix := range defaultPrefixes {
		pattern := `(?mi)^(\[|\()?` + prefix + `(\]|\))?(:|[[:blank:]])+`
		re := regexp.MustCompile(pattern)
		if re.FindString(title) != "" {
			return true
		}
	}
	return false
}

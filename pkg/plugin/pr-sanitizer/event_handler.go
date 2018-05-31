package prsanitizer

import (
	"encoding/json"
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
	ProwPluginName = "pr-sanitizer"

	// FailureMessage is a message used in GH Status as description when the PR title does not follow semantic message style
	FailureMessage = "PR title does not conform with semantic commit message style."

	// FailureDetailsLink is a link to an anchor in arq documentation that contains additional status details for InProgressMessage
	FailureDetailsLink = plugin.DocumentationURL + "#prsanitizer-failed"

	// SuccessMessage is a message used in GH Status as description when the PR title conforms to the semantic commit message style
	SuccessMessage = "PR title conforms with semantic commit message style."
	// SuccessDetailsLink is a link to an anchor in arq documentation that contains additional status details for ReadyForReviewMessage
	SuccessDetailsLink = plugin.DocumentationURL + "#prsanitizer-success"
)

// GitHubLabelsEventsHandler is the event handler for the plugin.
// Implements server.GitHubEventHandler interface which contains the logic for incoming GitHub events
type GitHubLabelsEventsHandler struct {
	Client  ghclient.Client
	BotName string
}

var (
	handledPrActions = []string{"opened", "reopened", "edited", "synchronized"}
	defaultTypes     = []string{"chore", "docs", "feat", "fix", "refactor", "style", "test"}
)

// HandleEvent is an entry point for the plugin logic. This method is invoked by the Server when
// events are dispatched from the /hook service
func (gh *GitHubLabelsEventsHandler) HandleEvent(log log.Logger, eventType github.EventType, payload []byte) error {
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
		if gh.HasSemanticMessage(*event.PullRequest.Title, configuration) {
			return statusService.Success(SuccessMessage, SuccessDetailsLink)
		}
		return statusService.Failure(FailureMessage, FailureDetailsLink)

	default:
		log.Warnf("received an event of type %q but didn't ask for it", eventType)
	}
	return nil
}

// HasSemanticMessage checks if title is marked as Work In Progress
func (gh *GitHubLabelsEventsHandler) HasSemanticMessage(title string, config PluginConfiguration) bool {
	prefixes := defaultTypes
	if len(config.TypePrefix) != 0 {
		if config.Combine {
			prefixes = append(prefixes, config.TypePrefix...)
		} else {
			prefixes = config.TypePrefix
		}
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(strings.ToLower(title), prefix) {
			return true
		}
	}
	return false
}

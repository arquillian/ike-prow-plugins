package prsanitizer

import (
	"encoding/json"
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/work-in-progress"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	gogh "github.com/google/go-github/github"
)

const (
	// ProwPluginName is an external prow plugin name used to register this service
	ProwPluginName = "pr-sanitizer"

	// TitleVerificationFailureMessage is a message used in GH Status as description when the PR title does not follow semantic message style
	TitleVerificationFailureMessage = "PR title does not conform with semantic commit message style."

	// TitleVerificationFailureDetailsPageName is a name of a documentation page that contains additional status details for title verification failure.
	TitleVerificationFailureDetailsPageName = "pr-sanitizer-failed"

	// SuccessMessage is a message used in GH Status as description when the PR title conforms to the semantic commit message style
	SuccessMessage = "PR title conforms with semantic commit message style."
	// SuccessDetailsPageName is a name of a documentation page that contains additional status details for success state
	SuccessDetailsPageName = "pr-sanitizer-success"
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

		if gh.HasTitleWithValidType(log, change, *event.PullRequest.Title) {
			return statusService.Success(SuccessMessage, SuccessDetailsPageName)
		}
		return statusService.Failure(TitleVerificationFailureMessage, TitleVerificationFailureDetailsPageName)

	default:
		log.Warnf("received an event of type %q but didn't ask for it", eventType)
	}
	return nil
}

// HasTitleWithValidType checks if title prefix conforms with semantic message style
func (gh *GitHubLabelsEventsHandler) HasTitleWithValidType(log log.Logger, change scm.RepositoryChange, title string) bool {
	title = gh.trimWorkInProgressPrefix(log, change, title)

	config := LoadConfiguration(log, change)
	prefixes := defaultTypes
	if len(config.TypePrefix) != 0 {
		if config.Combine {
			prefixes = append(prefixes, config.TypePrefix...)
		} else {
			prefixes = config.TypePrefix
		}
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(strings.ToLower(title), strings.ToLower(prefix)) {
			return true
		}
	}
	return false
}

func (gh *GitHubLabelsEventsHandler) trimWorkInProgressPrefix(log log.Logger, change scm.RepositoryChange, title string) string {
	wipPluginConfig := wip.LoadConfiguration(log, change)
	wipHandler := &wip.GitHubWIPPRHandler{Client: gh.Client, BotName: gh.BotName}

	if ok, prefix := wipHandler.HasWorkInProgressPrefix(title, wipPluginConfig); ok {
		title = strings.TrimSpace(strings.TrimPrefix(title, prefix))
	}
	return title
}

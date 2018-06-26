package prsanitizer

import (
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/command"
	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/work-in-progress"
	"github.com/arquillian/ike-prow-plugins/pkg/status"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	gogh "github.com/google/go-github/github"
)

const (
	// ProwPluginName is an external prow plugin name used to register this service
	ProwPluginName = "pr-sanitizer"

	// TitleVerificationFailureMessage is a message used in GH Status as description when the PR title does not follow semantic message style
	TitleVerificationFailureMessage = "PR title does not conform with semantic commit message style."

	// TitleVerificationFailureDetailsPageName is a name of a documentation page that contains additional status details for title verification failure.
	TitleVerificationFailureDetailsPageName = "title-verification-failed"

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
	handledCommentActions = []string{"created", "edited"}
	handledPrActions      = []string{"opened", "reopened", "edited", "synchronized"}
	defaultTypes          = []string{"chore", "docs", "feat", "fix", "refactor", "style", "test"}
)

// HandlePullRequestEvent is an entry point for the plugin logic. This method is invoked by the Server when
// pull request event is dispatched from the /hook service
func (gh *GitHubLabelsEventsHandler) HandlePullRequestEvent(log log.Logger, event *gogh.PullRequestEvent) error {
	if !utils.Contains(handledPrActions, *event.Action) {
		return nil
	}
	return gh.checkTitleAndSetStatus(log, event.PullRequest)
}

// HandleIssueCommentEvent is an entry point for the plugin logic. This method is invoked by the Server when
// issue comment event is dispatched from the /hook service
func (gh *GitHubLabelsEventsHandler) HandleIssueCommentEvent(log log.Logger, comment *gogh.IssueCommentEvent) error {
	if !utils.Contains(handledCommentActions, *comment.Action) {
		return nil
	}

	prLoader := ghservice.NewPullRequestLazyLoaderFromComment(gh.Client, comment)
	userPerm := command.NewPermissionService(gh.Client, *comment.Sender.Login, prLoader)

	cmdHandler := command.CommentCmdHandler{Client: gh.Client}
	cmdHandler.Register(&command.RunCmd{
		PluginName:            ProwPluginName,
		UserPermissionService: userPerm,
		WhenAddedOrEdited: func() error {
			pullRequest, err := prLoader.Load()
			if err != nil {
				return err
			}

			return gh.checkTitleAndSetStatus(log, pullRequest)
		}})

	err := cmdHandler.Handle(log, comment)
	if err != nil {
		log.Error(err)
	}
	return err
}

func (gh *GitHubLabelsEventsHandler) checkTitleAndSetStatus(log log.Logger, pullRequest *gogh.PullRequest) error {
	change := ghservice.NewRepositoryChangeForPR(pullRequest)
	statusContext := github.StatusContext{BotName: gh.BotName, PluginName: ProwPluginName}
	statusService := status.NewStatusService(gh.Client, log, change, statusContext)

	config := LoadConfiguration(log, change)
	if gh.HasTitleWithValidType(config, *pullRequest.Title) {
		return statusService.Success(SuccessMessage, SuccessDetailsPageName)
	} else if prefix, ok := wip.GetWorkInProgressPrefix(*pullRequest.Title, wip.LoadConfiguration(log, change)); ok {
		trimmedTitle := strings.TrimPrefix(*pullRequest.Title, prefix)
		if gh.HasTitleWithValidType(config, trimmedTitle) {
			return statusService.Success(SuccessMessage, SuccessDetailsPageName)
		}
	}
	return statusService.Failure(TitleVerificationFailureMessage, TitleVerificationFailureDetailsPageName)
}

// HasTitleWithValidType checks if title prefix conforms with semantic message style
func (gh *GitHubLabelsEventsHandler) HasTitleWithValidType(config PluginConfiguration, title string) bool {
	prefixes := defaultTypes
	if len(config.TypePrefix) != 0 {
		if config.Combine {
			prefixes = append(prefixes, config.TypePrefix...)
		} else {
			prefixes = config.TypePrefix
		}
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(title)), strings.ToLower(strings.TrimSpace(prefix))) {
			return true
		}
	}
	return false
}

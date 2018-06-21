package wip

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/command"
	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
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
	handledCommentActions = []string{"created", "edited"}
	handledPrActions      = []string{"opened", "reopened", "edited", "synchronize", "labeled", "unlabeled"}
	defaultPrefixes       = []string{"WIP", "DO NOT MERGE", "DON'T MERGE", "WORK-IN-PROGRESS"}
)

// HandlePullRequestEvent is an entry point for the plugin logic. This method is invoked by the Server when
// Pull Request Event is dispatched from the /hook service
func (gh *GitHubWIPPRHandler) HandlePullRequestEvent(log log.Logger, event *gogh.PullRequestEvent) error {
	if !utils.Contains(handledPrActions, *event.Action) {
		return nil
	}

	switch *event.Action {
	case github.ActionLabeled, github.ActionUnlabeled:
		return gh.checkComponentsAndSetStatus(log, event.PullRequest, true)
	default:
		return gh.checkComponentsAndSetStatus(log, event.PullRequest, false)
	}
}

// HandleIssueCommentEvent is an entry point for the plugin logic. This method is invoked by the Server when
// Issue Comment Event is dispatched from the /hook service
func (gh *GitHubWIPPRHandler) HandleIssueCommentEvent(log log.Logger, comment *gogh.IssueCommentEvent) error {
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

			return gh.checkComponentsAndSetStatus(log, pullRequest, false)

		}})

	err := cmdHandler.Handle(log, comment)
	if err != nil {
		log.Error(err)
	}
	return err
}

func (gh *GitHubWIPPRHandler) checkComponentsAndSetStatus(log log.Logger, pullRequest *gogh.PullRequest, labelUpdated bool) error {
	change := ghservice.NewRepositoryChangeForPR(pullRequest)
	statusContext := github.StatusContext{BotName: gh.BotName, PluginName: ProwPluginName}
	statusService := ghservice.NewStatusService(gh.Client, log, change, statusContext)

	configuration := LoadConfiguration(log, change)
	labelExists := gh.hasWorkInProgressLabel(pullRequest.Labels, configuration.Label)
	prefix, prefixExists := GetWorkInProgressPrefix(*pullRequest.Title, configuration)

	if prefixExists && !labelExists {
		if labelUpdated {
			*pullRequest.Title = strings.TrimSpace(strings.TrimPrefix(*pullRequest.Title, prefix))
			if err := gh.Client.EditPullRequest(pullRequest); err != nil {
				return fmt.Errorf("failed to update PR title [%q]. cause: %s", *pullRequest, err)
			}
			return statusService.Success(ReadyForReviewMessage, ReadyForReviewDetailsPageName)
		}
		if err := gh.Client.AddPullRequestLabel(change, *pullRequest.Number, []string{configuration.Label}); err != nil {
			log.Errorf("failed to add label on PR [%q]. cause: %s", *pullRequest, err)
		}
		return statusService.Failure(InProgressMessage, InProgressDetailsPageName)
	}
	if labelExists {
		if !prefixExists && !labelUpdated {
			if err := gh.Client.RemovePullRequestLabel(change, *pullRequest.Number, configuration.Label); err != nil {
				log.Errorf("failed to remove label on PR [%q]. cause: %s", *pullRequest, err)
			}
			return statusService.Success(ReadyForReviewMessage, ReadyForReviewDetailsPageName)
		}
		return statusService.Failure(InProgressMessage, InProgressDetailsPageName)
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

// GetWorkInProgressPrefix checks if title is marked as Work In Progress
func GetWorkInProgressPrefix(title string, config PluginConfiguration) (string, bool) {
	prefixes := defaultPrefixes
	if len(config.Prefix) != 0 {
		if config.Combine {
			prefixes = append(prefixes, config.Prefix...)
		} else {
			prefixes = config.Prefix
		}
	}
	return hasPrefix(title, prefixes)
}

func hasPrefix(title string, prefixes []string) (string, bool) {
	for _, prefix := range prefixes {
		pattern := `(?mi)^(\[|\()?` + prefix + `(\]|\))?(:| )+`
		r, _ := regexp.Compile(pattern)
		if match := r.FindString(title); match != "" {
			return strings.TrimSpace(match), true
		}
	}
	return "", false
}

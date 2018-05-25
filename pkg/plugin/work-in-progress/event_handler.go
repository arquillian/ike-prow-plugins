package wip

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
	"github.com/arquillian/ike-prow-plugins/pkg/command"
)

const (
	// ProwPluginName is an external prow plugin name used to register this service
	ProwPluginName = "work-in-progress"

	// WipPrefix is a prefix which, when applied on the PR title, marks its state as "work-in-progress"
	WipPrefix = "wip "

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
	handledPrActions      = []string{"opened", "reopened", "edited"}
	handledCommentActions = []string{"created"}
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
		if gh.IsWorkInProgress(event.PullRequest.Title) {
			return statusService.Failure(InProgressMessage, InProgressDetailsLink)
		}
		return statusService.Success(ReadyForReviewMessage, ReadyForReviewDetailsLink)

	case github.IssueComment:
		var event gogh.IssueCommentEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			log.Errorf("Failed while parsing '%q' event with payload: %q. Cause: %q", github.IssueComment, event, err)
			return err
		}

		if err := gh.handlePrComment(log, &event); err != nil {
			log.Errorf("Error handling '%q' event with payload %q. Cause: %q", github.IssueComment, event, err)
			return err
		}

	default:
		log.Warnf("received an event of type %q but didn't ask for it", eventType)
	}

	return nil
}

// IsWorkInProgress checks if title is marked as Work In Progress
func (gh *GitHubWIPPRHandler) IsWorkInProgress(title *string) bool {
	return strings.HasPrefix(strings.ToLower(*title), WipPrefix)
}

func (gh *GitHubWIPPRHandler) handlePrComment(log log.Logger, comment *gogh.IssueCommentEvent) error {
	if !utils.Contains(handledCommentActions, *comment.Action) {
		return nil
	}

	prLoader := ghservice.NewPullRequestLazyLoader(gh.Client, comment)
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

			change := scm.RepositoryChange{
				Owner:    *comment.Repo.Owner.Login,
				RepoName: *comment.Repo.Name,
				Hash:     *pullRequest.Head.SHA,
			}

			statusContext := github.StatusContext{BotName: gh.BotName, PluginName: ProwPluginName}
			statusService := ghservice.NewStatusService(gh.Client, log, change, statusContext)
			if gh.IsWorkInProgress(pullRequest.Title) {
				return statusService.Failure(InProgressMessage, InProgressDetailsLink)
			}
			return statusService.Success(ReadyForReviewMessage, ReadyForReviewDetailsLink)
		}})

	err := cmdHandler.Handle(log, comment)
	if err != nil {
		log.Error(err)
	}
	return err
}

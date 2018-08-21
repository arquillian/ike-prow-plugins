package prsanitizer

import (
	"github.com/arquillian/ike-prow-plugins/pkg/command"
	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	gogh "github.com/google/go-github/github"
)

// GitHubPRSanitizerEventsHandler is the event handler for the plugin.
// Implements server.GitHubEventHandler interface which contains the logic for incoming GitHub events
type GitHubPRSanitizerEventsHandler struct {
	Client  ghclient.Client
	BotName string
}

var (
	handledCommentActions = []string{"created", "edited"}
	handledPrActions      = []string{"opened", "reopened", "edited", "synchronize"}
)

const documentationSection = "#_pr_sanitizer_plugin"

// HandlePullRequestEvent is an entry point for the plugin logic. This method is invoked by the Server when
// pull request event is dispatched from the /hook service
func (gh *GitHubPRSanitizerEventsHandler) HandlePullRequestEvent(log log.Logger, event *gogh.PullRequestEvent) error {
	if !utils.Contains(handledPrActions, *event.Action) {
		return nil
	}
	return gh.validatePullRequestTitleAndDescription(log, event.PullRequest)
}

// HandleIssueCommentEvent is an entry point for the plugin logic. This method is invoked by the Server when
// issue comment event is dispatched from the /hook service
func (gh *GitHubPRSanitizerEventsHandler) HandleIssueCommentEvent(log log.Logger, comment *gogh.IssueCommentEvent) error {
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

			return gh.validatePullRequestTitleAndDescription(log, pullRequest)
		}})

	err := cmdHandler.Handle(log, comment)
	if err != nil {
		log.Error(err)
	}
	return err
}

func (gh *GitHubPRSanitizerEventsHandler) validatePullRequestTitleAndDescription(log log.Logger, pr *gogh.PullRequest) error {
	change := ghservice.NewRepositoryChangeForPR(pr)
	config := LoadConfiguration(log, change)
	statusService := gh.newPrSanitizerStatusService(log, pr, config)

	messages := executeChecks(pr, config, log)

	if len(messages) > 0 {
		return statusService.fail(messages)
	}

	return statusService.success()
}

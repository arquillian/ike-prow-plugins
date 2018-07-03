package prsanitizer

import (
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/command"
	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/work-in-progress"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	gogh "github.com/google/go-github/github"
	"regexp"
	"github.com/arquillian/ike-prow-plugins/pkg/status/message"
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
	defaultTypes          = []string{"chore", "docs", "feat", "fix", "refactor", "style", "test"}
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
	statusService := gh.newPRTitleDescriptionStatusService(log, change)
	config := LoadConfiguration(log, change)
	isTitleWithValidType := gh.HasTitleWithValidType(config, *pr.Title)
	if !isTitleWithValidType {
		if prefix, ok := wip.GetWorkInProgressPrefix(*pr.Title, wip.LoadConfiguration(log, change)); ok {
			trimmedTitle := strings.TrimPrefix(*pr.Title, prefix)
			isTitleWithValidType = gh.HasTitleWithValidType(config, trimmedTitle)
		}
	}

	description, isIssueLinked := gh.GetDescriptionWithIssueLinkExcluded(pr.GetBody())

	failureMessageBuilder := NewFailureMessageBuilder()
	hintMessage := failureMessageBuilder.Title(isTitleWithValidType).Description(description, config.DescriptionContentLength).IssueLink(isIssueLinked).Build()

	if len(hintMessage) > 0 {
		commentsLoader := ghservice.NewIssueCommentsLazyLoader(gh.Client, pr)
		msgContext := message.NewStatusMessageContext(ProwPluginName, documentationSection, pr, config.PluginConfiguration)
		msgService := message.NewStatusMessageService(gh.Client, log, commentsLoader, msgContext)
		msgService.SadStatusMessageForPRSanitizer(string(hintMessage), true)

		return statusService.fail()
	}

	return statusService.titleAndDescriptionOk()
}

// GetDescriptionWithIssueLinkExcluded return description with excluding issue link keyword.
func (gh *GitHubPRSanitizerEventsHandler) GetDescriptionWithIssueLinkExcluded(description string) (string, bool) {
	desc := strings.ToLower(description)
	var issueLink = regexp.MustCompile(`(close|closes|closed|fix|fixes|fixed|resolve|resolves|resolved)[:]?[\s]+[\w-/]*#[\d]+`)
	return strings.TrimSpace(issueLink.ReplaceAllString(desc, "")), issueLink.MatchString(desc)
}

// HasTitleWithValidType checks if title prefix conforms with semantic message style.
func (gh *GitHubPRSanitizerEventsHandler) HasTitleWithValidType(config PluginConfiguration, title string) bool {
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

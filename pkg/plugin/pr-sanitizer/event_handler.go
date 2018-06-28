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
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"fmt"
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

// DescriptionShortMessage is message notification for contributor about PR short description content.
const DescriptionShortMessage = "Hey @%s! It seems that PR description is too short. More elaborated description will be helpful to " +
	"understand changes proposed in this PR."

// HandlePullRequestEvent is an entry point for the plugin logic. This method is invoked by the Server when
// pull request event is dispatched from the /hook service
func (gh *GitHubLabelsEventsHandler) HandlePullRequestEvent(log log.Logger, event *gogh.PullRequestEvent) error {
	if !utils.Contains(handledPrActions, *event.Action) {
		return nil
	}
	return gh.checkTitleDescriptionAndSetStatus(log, event.PullRequest)
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

			return gh.checkTitleDescriptionAndSetStatus(log, pullRequest)
		}})

	err := cmdHandler.Handle(log, comment)
	if err != nil {
		log.Error(err)
	}
	return err
}

func (gh *GitHubLabelsEventsHandler) checkTitleDescriptionAndSetStatus(log log.Logger, pr *gogh.PullRequest) error {
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

	if len(description) < 50 {
		message := fmt.Sprintf(DescriptionShortMessage, *pr.User.Login)
		err := gh.Client.CreateIssueComment(scm.RepositoryIssue{
			Owner:    *pr.Base.Repo.Owner.Login,
			RepoName: *pr.Base.Repo.Name,
			Number:   *pr.Number,
		}, &message)
		if err != nil {
			log.Errorf("failed to comment on PR [%q]. cause: %s", *pr, err)
			return err
		}
	}

	failureMessageBuilder := NewFailureMessageBuilder()
	failureMessage := failureMessageBuilder.Title(isTitleWithValidType).Description(description).IssueLink(isIssueLinked).Build()

	if len(failureMessage) > 0 {
		return statusService.fail(failureMessage)
	}

	return statusService.titleAndDescriptionOk()
}

// GetDescriptionWithIssueLinkExcluded return description with excluding issue link keyword.
func (gh *GitHubLabelsEventsHandler) GetDescriptionWithIssueLinkExcluded(d string) (string, bool) {
	desc := strings.ToLower(d)
	var issueLink = regexp.MustCompile(`(close|closes|closed|fix|fixes|fixed|resolve|resolves|resolved)[:]?[\s]+[\w-/]*#[\d]+`)
	return strings.TrimSpace(issueLink.ReplaceAllString(desc, "")), issueLink.MatchString(desc)
}

// HasTitleWithValidType checks if title prefix conforms with semantic message style.
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

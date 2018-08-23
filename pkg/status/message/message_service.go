package message

import (
	"fmt"
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/config"
	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	gogh "github.com/google/go-github/github"
)

// PluginTitleTemplate is a constant template containing "Ike Plugins (name-of-plugin)" title with markdown formatting
const (
	PluginTitleTemplate     = "### Ike Plugins (%s)"
	assigneeMentionTemplate = "Thank you @%s for this contribution!"

	sadIke   = `<img align="left" src="https://raw.githubusercontent.com/arquillian/ike-prow-plugins/master/docs/images/arquillian_ui_failure_64px.png">`
	happyIke = `<img align="left" src="https://raw.githubusercontent.com/arquillian/ike-prow-plugins/master/docs/images/arquillian_ui_success_64px.png">`
)

// StatusMessageService is a struct managing plugin comments
type StatusMessageService struct {
	commentService *ghservice.CommentService
	log            log.Logger
	commentContext StatusMessageContext
	commentsLoader *ghservice.IssueCommentsLazyLoader
	change         scm.RepositoryChange
}

// StatusMessageContext holds a plugin name and a assignee to be mentioned in the comment
type StatusMessageContext struct {
	pluginName           string
	documentationSection string
	pullRequest          *gogh.PullRequest
	config               config.PluginConfiguration
}

// NewStatusMessageContext creates an instance of StatusMessageContext with the given values
func NewStatusMessageContext(pluginName, documentationSection string, pr *gogh.PullRequest, config config.PluginConfiguration) StatusMessageContext {
	return StatusMessageContext{
		pluginName:           pluginName,
		documentationSection: documentationSection,
		pullRequest:          pr,
		config:               config,
	}
}

// NewStatusMessageService creates an instance of GitHub StatusMessageService for the given StatusMessageContext
func NewStatusMessageService(client ghclient.Client, log log.Logger, commentsLoader *ghservice.IssueCommentsLazyLoader, commentContext StatusMessageContext) *StatusMessageService {
	return &StatusMessageService{
		commentService: &ghservice.CommentService{
			Client: client,
			Issue:  commentsLoader.Issue,
		},
		log:            log,
		commentContext: commentContext,
		commentsLoader: commentsLoader,
		change:         ghservice.NewRepositoryChangeForPR(commentContext.pullRequest),
	}
}

// SadStatusMessage creates a message with the sad Ike image
func (s *StatusMessageService) SadStatusMessage(description, statusFileSpec string, addIfMissing bool) {
	s.logError(s.StatusMessage(func() string {
		messageLoader := s.newMessageLoader(sadIke, description)
		return messageLoader.LoadMessage(s.change, statusFileSpec)
	}, addIfMissing))
}

// HappyStatusMessage creates a message with the happy Ike image
func (s *StatusMessageService) HappyStatusMessage(description, statusFileSpec string, addIfMissing bool) {
	s.logError(s.StatusMessage(func() string {
		messageLoader := s.newMessageLoader(happyIke, description)
		return messageLoader.LoadMessage(s.change, statusFileSpec)
	}, addIfMissing))
}

func (s *StatusMessageService) logError(err error) {
	if err != nil {
		s.log.Errorf("failed to comment on PR, caused by: %s", err)
	}
}

func (s *StatusMessageService) newMessageLoader(image, msg string) *Loader {
	return &Loader{
		Log:        s.log,
		PluginName: s.commentContext.pluginName,
		Message: &Message{
			Thumbnail:     image,
			Description:   msg,
			ConfigFile:    s.commentContext.config.LocationURL,
			Documentation: s.commentContext.documentationSection,
		},
	}
}

// StatusMessage checks all present comments in the issue/pull-request. If no comment with PluginTitleTemplate
// (with the related plugin) is found, then it adds a new comment with the plugin title, assignee mention
// and the given commentMsg. If such a comment is present already, then it does nothing.
func (s *StatusMessageService) StatusMessage(commentMsgCreator func() string, addIfMissing bool) error {
	comments, err := s.commentsLoader.Load()
	if err != nil {
		s.log.Errorf("Getting all comments failed with an error: %s", err)
	}
	statusMsg := utils.String("")

	for _, com := range comments {
		content := *com.Body
		if strings.HasPrefix(content, s.getPluginTitle()) {
			s.loadStatusMessage(commentMsgCreator, statusMsg)
			if strings.TrimSpace(content) == strings.TrimSpace(*statusMsg) {
				return nil
			}
			return s.commentService.EditComment(*com.ID, statusMsg)
		}
	}
	if addIfMissing {
		s.loadStatusMessage(commentMsgCreator, statusMsg)
		return s.commentService.AddComment(statusMsg)
	}
	return nil
}

func (s *StatusMessageService) append(first, second string) string {
	return first + "\n\n" + second
}

func (s *StatusMessageService) createPluginStatusMsg(commentMsg string) *string {
	return utils.String(s.append(s.createBeginning(), commentMsg))
}

func (s *StatusMessageService) createBeginning() string {
	return s.append(s.getPluginTitle(), fmt.Sprintf(assigneeMentionTemplate, *s.commentContext.pullRequest.User.Login))
}

func (s *StatusMessageService) getPluginTitle() string {
	return fmt.Sprintf(PluginTitleTemplate, s.commentContext.pluginName)
}

func (s *StatusMessageService) loadStatusMessage(commentMsgCreator func() string, statusMsg *string) {
	if *statusMsg == "" {
		*statusMsg = *s.createPluginStatusMsg(commentMsgCreator())
	}
}

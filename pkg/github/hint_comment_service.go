package github

import (
	"fmt"
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
)

// PluginTitleTemplate is a constant template containing "Ike Plugins (name-of-plugin)" title with markdown formatting
const (
	PluginTitleTemplate     = "### Ike Plugins (%s)"
	assigneeMentionTemplate = "Thank you @%s for this contribution!"
)

// HintCommentService is a struct managing plugin comments
type HintCommentService struct {
	*CommentService
	log            log.Logger
	commentContext HintCommentContext
}

// HintCommentContext holds a plugin name and a assignee to be mentioned in the comment
type HintCommentContext struct {
	PluginName string
	Assignee   string // TODO rethink this naming when plugins will start interacting with issue creators and reviewers
}

// NewHintCommentService creates an instance of GitHub HintCommentService for the given HintCommentContext
func NewHintCommentService(client *Client, log log.Logger, change scm.RepositoryChange, issueOrPrNumber int, commentContext HintCommentContext) *HintCommentService {
	return &HintCommentService{
		CommentService: &CommentService{
			client: client,
			issue: scm.RepositoryIssue{
				Owner:    change.Owner,
				RepoName: change.RepoName,
				Number:   issueOrPrNumber,
			},
		},
		log:            log,
		commentContext: commentContext,
	}
}

// PluginComment checks all present comments in the issue/pull-request. If no comment with PluginTitleTemplate
// (with the related plugin) is found, then it adds a new comment with the plugin title, assignee mention
// and the given commentMsg. If such a comment is present already, then it does nothing.
func (s *HintCommentService) PluginComment(commentMsg string) error {

	comments, err := s.client.ListIssueComments(s.issue)
	if err != nil {
		s.log.Errorf("Getting all comments failed with an error: %s", err)
	}

	for _, com := range comments {
		content := *com.Body
		if strings.HasPrefix(content, s.getPluginTitle()) {
			return nil
		}
	}

	return s.Comment(s.createPluginHint(commentMsg))
}

func (s *HintCommentService) append(first, second string) string {
	return first + "\n\n" + second
}

func (s *HintCommentService) createPluginHint(commentMsg string) *string {
	return utils.String(s.append(s.createBeginning(), commentMsg))
}

func (s *HintCommentService) createBeginning() string {
	return s.append(s.getPluginTitle(), fmt.Sprintf(assigneeMentionTemplate, s.commentContext.Assignee))
}

func (s *HintCommentService) getPluginTitle() string {
	return fmt.Sprintf(PluginTitleTemplate, s.commentContext.PluginName)
}

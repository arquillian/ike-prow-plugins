package plugin

import (
	"context"
	"fmt"
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
)

// PluginTitleTemplate is a constant template containing "Ike Plugins (name-of-plugin)" title with markdown formatting
const (
	PluginTitleTemplate     = "### Ike Plugins (%s)"
	assigneeMentionTemplate = "@%s Please, pay attention to the following message:"
)

// CommentService is a struct managing plugin comments
type CommentService struct {
	client         *github.Client
	log            *logrus.Entry
	issue          scm.RepositoryIssue
	commentContext CommentContext
}

// CommentContext holds a plugin name and a assignee to be mentioned in the comment
type CommentContext struct {
	PluginName string
	Assignee   string // TODO rethink this naming when plugins will start interacting with issue creators and reviewers
}

// NewCommentService creates an instance of GitHub CommentService for the given CommentContext
func NewCommentService(client *github.Client, log *logrus.Entry, change scm.RepositoryChange, issueOrPrNumber int, commentContext CommentContext) *CommentService {
	return &CommentService{
		client: client,
		log:    log,
		issue: scm.RepositoryIssue{
			Owner:    change.Owner,
			RepoName: change.RepoName,
			Number:   issueOrPrNumber,
		},
		commentContext: commentContext,
	}
}

// PluginComment checks all present comments in the issue/pull-request. If no comment with PluginTitleTemplate
// (with the related plugin) is found, then it adds a new comment with the plugin title, assignee mention
// and the given commentMsg. If such a comment is present already, then it does nothing.
func (s *CommentService) PluginComment(commentMsg string) error {

	comments, _, err := s.client.Issues.
		ListComments(context.Background(), s.issue.Owner, s.issue.RepoName, s.issue.Number, &github.IssueListCommentsOptions{})
	if err != nil {
		s.log.Errorf("Getting all comments failed with an error: %s", err)
	}

	for _, com := range comments {
		content := *com.Body
		if strings.HasPrefix(content, s.getPluginTitle()) {
			return nil
		}
	}

	comment := &github.IssueComment{
		Body: s.createPluginHint(commentMsg),
	}

	_, _, err = s.client.Issues.CreateComment(context.Background(), s.issue.Owner, s.issue.RepoName, s.issue.Number, comment)

	return err
}

func (s *CommentService) append(first, second string) string {
	return first + "\n\n" + second
}

func (s *CommentService) createPluginHint(commentMsg string) *string {
	return utils.String(s.append(s.createBeginning(), commentMsg))
}

func (s *CommentService) createBeginning() string {
	return s.append(s.getPluginTitle(), fmt.Sprintf(assigneeMentionTemplate, s.commentContext.Assignee))
}

func (s *CommentService) getPluginTitle() string {
	return fmt.Sprintf(PluginTitleTemplate, s.commentContext.PluginName)
}

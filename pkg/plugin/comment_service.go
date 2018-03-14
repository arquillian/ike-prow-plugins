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

// IkePluginsTitle is a constant containing "Ike Plugins" title with markdown formatting
const IkePluginsTitle = "### Ike Plugins"

// CommentService is a struct managing plugin comments
type CommentService struct {
	client         *github.Client
	log            *logrus.Entry
	issue          RepositoryIssue
	commentContext CommentContext
}

// RepositoryIssue holds owner name, repository name and an issue number
type RepositoryIssue struct {
	Owner    string
	RepoName string
	Number   int
}

// CommentContext holds a plugin name and a assignee to be mentioned in the comment
type CommentContext struct {
	PluginName string
	Assignee   string
}

// NewCommentService creates an instance of GitHub CommentService for the given CommentContext
func NewCommentService(client *github.Client, log *logrus.Entry, change scm.RepositoryChange, issueOrPrNumber int, commentContext CommentContext) *CommentService {
	return &CommentService{
		client: client,
		log:    log,
		issue: RepositoryIssue{
			Owner:    change.Owner,
			RepoName: change.RepoName,
			Number:   issueOrPrNumber,
		},
		commentContext: commentContext,
	}
}

// PluginComment checks all present comments in the issue/pull-request. If no comment with IkePluginsTitle is found, then
// it adds a new comment with IkePluginsTitle, assignee, the plugin tittle and the given commentMsg. If a comment with
// IkePluginsTitle is found but missing the particular plugin title, then it modifies the comment by adding plugin title
// with the given commentMsg. If everything is present already, then it does nothing.
func (s *CommentService) PluginComment(commentMsg string) error {

	comments, _, err := s.client.Issues.
		ListComments(context.Background(), s.issue.Owner, s.issue.RepoName, s.issue.Number, &github.IssueListCommentsOptions{})
	if err != nil {
		s.log.Errorf("Getting all comments failed with an error: %s", err)
	}

	for _, com := range comments {
		content := *com.Body
		if strings.HasPrefix(content, IkePluginsTitle) {
			if strings.Contains(content, s.getPluginTitle()) {
				return nil
			}
			return s.appendToComment(commentMsg, com)
		}
	}

	comment := &github.IssueComment{
		Body: s.append(s.createBeginning(), s.createPluginParagraph(commentMsg)),
	}

	_, _, err = s.client.Issues.CreateComment(context.Background(), s.issue.Owner, s.issue.RepoName, s.issue.Number, comment)

	return err
}

func (s *CommentService) appendToComment(commentMsg string, comment *github.IssueComment) error {
	comment.Body = s.append(*comment.Body, s.createPluginParagraph(commentMsg))
	_, _, err := s.client.Issues.EditComment(context.Background(), s.issue.Owner, s.issue.RepoName, int(*comment.ID), comment)
	return err
}

func (s *CommentService) append(first, second string) *string {
	return utils.String(first + "\n\n" + second)
}

func (s *CommentService) createPluginParagraph(commentMsg string) string {
	return s.getPluginTitle() + "\n\n" + commentMsg
}

func (s *CommentService) createBeginning() string {
	return *s.append(IkePluginsTitle, fmt.Sprintf("@%s Please, pay attention to the following message:", s.commentContext.Assignee))
}

func (s *CommentService) getPluginTitle() string {
	return "#### " + s.commentContext.PluginName
}

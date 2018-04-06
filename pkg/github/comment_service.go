package github

import (
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	gogh "github.com/google/go-github/github"
)

// CommentService is a struct managing issue or pull request comments
type CommentService struct {
	client *Client
	issue  scm.RepositoryIssue
}

// NewCommentService creates an instance of GitHub CommentService with information retrieved from the given IssueCommentEvents
func NewCommentService(client *Client, prComment *gogh.IssueCommentEvent) *CommentService {
	return &CommentService{
		client: client,
		issue: scm.RepositoryIssue{
			Owner:    *prComment.Repo.Owner.Login,
			RepoName: *prComment.Repo.Name,
			Number:   *prComment.Issue.Number,
		},
	}
}

// AddComment adds a comment message to the issue
func (s *CommentService) AddComment(commentMsg *string) error {
	return s.client.CreateIssueComment(s.issue, commentMsg)
}

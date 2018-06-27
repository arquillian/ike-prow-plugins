package ghservice

import (
	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	gogh "github.com/google/go-github/github"
)

// CommentService is a struct managing issue or pull request comments
type CommentService struct {
	Client ghclient.Client
	Issue  scm.RepositoryIssue
}

// NewCommentService creates an instance of GitHub CommentService with information retrieved from the given IssueCommentEvents
func NewCommentService(client ghclient.Client, comment *gogh.IssueCommentEvent) *CommentService {
	return &CommentService{
		Client: client,
		Issue: scm.RepositoryIssue{
			Owner:    *comment.Repo.Owner.Login,
			RepoName: *comment.Repo.Name,
			Number:   *comment.Issue.Number,
		},
	}
}

// AddComment adds a comment message to the issue
func (s *CommentService) AddComment(commentMsg *string) error {
	return s.Client.CreateIssueComment(s.Issue, commentMsg)
}

// EditComment edits an already existing comment message
func (s *CommentService) EditComment(commentID int64, commentMsg *string) error {
	return s.Client.EditIssueComment(s.Issue, commentID, commentMsg)
}

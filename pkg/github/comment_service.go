package github

import (
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
)

// CommentService is a struct managing issue or pull request comments
type CommentService struct {
	client *Client
	issue  scm.RepositoryIssue
}

// Comment adds a comment message to the issue
func (s *CommentService) Comment(commentMsg *string) error {
	return s.client.CreateIssueComment(s.issue, commentMsg)
}

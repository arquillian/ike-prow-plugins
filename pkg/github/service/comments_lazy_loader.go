package ghservice

import (
	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	gogh "github.com/google/go-github/github"
)

// IssueCommentsLazyLoader represents a lazy loader of issue comments - is loaded when needed and only once
type IssueCommentsLazyLoader struct {
	Client        ghclient.Client
	Issue         scm.RepositoryIssue
	issueComments []*gogh.IssueComment
	err           error
}

// NewIssueCommentsLazyLoader creates a new instance of IssueCommentsLazyLoader with information retrieved from the given pull request
func NewIssueCommentsLazyLoader(client ghclient.Client, pr *gogh.PullRequest) *IssueCommentsLazyLoader {
	return &IssueCommentsLazyLoader{
		Client: client,
		Issue:  *scm.NewRepositoryIssue(*pr.Base.Repo.Owner.Login, *pr.Base.Repo.Name, *pr.Number),
	}
}

// Load loads list of issue comments - if not already retrieved from GH then it gets it and stores, then it uses
// this stored instance for every future call
func (r *IssueCommentsLazyLoader) Load() ([]*gogh.IssueComment, error) {
	if r.issueComments == nil {
		r.issueComments, r.err = r.Client.ListIssueComments(r.Issue)
	}
	return r.issueComments, r.err
}

package ghservice

import (
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	gogh "github.com/google/go-github/v41/github"
)

// NewRepositoryChangeForPR creates a RepositoryChange instance for hte given pull request
func NewRepositoryChangeForPR(pr *gogh.PullRequest) scm.RepositoryChange {
	return scm.RepositoryChange{
		Owner:    *pr.Base.Repo.Owner.Login,
		RepoName: *pr.Base.Repo.Name,
		Hash:     *pr.Head.SHA,
	}
}

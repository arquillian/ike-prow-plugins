package github

import (
	"context"

	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	gogh "github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type client struct {
	Client *gogh.Client
}

// Client manages communication with the GitHub API.
type Client interface {
	GetPermissionLevel(owner, repo, user string) (*gogh.RepositoryPermissionLevel, error)
	GetPullRequest(owner, repo string, prNumber int) (*gogh.PullRequest, error)
	ListPullRequestFiles(owner, repo string, prNumber int) ([]scm.ChangedFile, error)
	ListIssueComments(issue scm.RepositoryIssue) ([]*gogh.IssueComment, error)
	CreateIssueComment(issue scm.RepositoryIssue, commentMsg *string) error
	CreateStatus(change scm.RepositoryChange, repoStatus *gogh.RepoStatus) error
	// This method is intended to be used by client decrators which need access to GH API methods not (yet)
	// exposed by this interface
	unwrap() *gogh.Client
}

// NewOauthClient creates a Client instance with the given oauth secret used as a access token. Underneath
// it creates go-github client which is used as delegate
func NewOauthClient(oauthSecret []byte) Client {
	token := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: string(oauthSecret)})
	oauthClient := gogh.NewClient(oauth2.NewClient(context.Background(), token))
	return NewClient(oauthClient)
}

// NewClient creates a Client instance with the given instance of go-github client which will be used as a delegate
func NewClient(c *gogh.Client) Client {
	return &client{c}
}

// GetPermissionLevel retrieves the specific permission level a collaborator has for a given repository.
func (c client) GetPermissionLevel(owner, repo, user string) (*gogh.RepositoryPermissionLevel, error) {
	permissionLevel, response, err := c.Client.Repositories.GetPermissionLevel(context.Background(), owner, repo, user)
	return permissionLevel, responseFailureCodeOrErr(response, err)
}

// GetPullRequest retrieves information about a single pull request.
func (c client) GetPullRequest(owner, repo string, prNumber int) (*gogh.PullRequest, error) {
	pr, response, err := c.Client.PullRequests.Get(context.Background(), owner, repo, prNumber)

	return pr, responseFailureCodeOrErr(response, err)
}

// ListPullRequestFiles lists the changed files in a pull request.
func (c client) ListPullRequestFiles(owner, repo string, prNumber int) ([]scm.ChangedFile, error) {
	files, response, err := c.Client.PullRequests.ListFiles(context.Background(), owner, repo, prNumber, nil)
	changedFiles := make([]scm.ChangedFile, 0, len(files))
	for _, file := range files {
		changedFiles = append(changedFiles, *scm.NewChangedFile(file))
	}
	return changedFiles, responseFailureCodeOrErr(response, err)
}

// ListIssueComments lists all comments on the specified issue.
func (c client) ListIssueComments(issue scm.RepositoryIssue) ([]*gogh.IssueComment, error) {
	comments, response, err :=
		c.Client.Issues.ListComments(context.Background(), issue.Owner, issue.RepoName, issue.Number, &gogh.IssueListCommentsOptions{})

	return comments, responseFailureCodeOrErr(response, err)
}

// CreateIssueComment creates a new comment on the specified issue.
func (c client) CreateIssueComment(issue scm.RepositoryIssue, commentMsg *string) error {
	comment := &gogh.IssueComment{
		Body: commentMsg,
	}
	_, response, err := c.Client.Issues.CreateComment(context.Background(), issue.Owner, issue.RepoName, issue.Number, comment)
	return responseFailureCodeOrErr(response, err)
}

// CreateStatus creates a new status for a repository at the specified reference represented by a RepositoryChange
func (c client) CreateStatus(change scm.RepositoryChange, repoStatus *gogh.RepoStatus) error {
	_, response, err :=
		c.Client.Repositories.CreateStatus(context.Background(), change.Owner, change.RepoName, change.Hash, repoStatus)
	return responseFailureCodeOrErr(response, err)
}

func (c client) unwrap() *gogh.Client {
	return c.Client
}

func responseFailureCodeOrErr(response *gogh.Response, e error) error {
	if response != nil && response.StatusCode >= 404 {
		return fmt.Errorf("server responded with error %d", response.StatusCode)
	}
	return e
}

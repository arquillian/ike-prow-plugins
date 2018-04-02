package github

import (
	"context"

	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	gogh "github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"time"
	"github.com/arquillian/ike-prow-plugins/pkg/http"
)

// Client manages communication with the GitHub API.
type Client struct {
	Client  *gogh.Client
	Retries int
	Sleep   time.Duration
}

// NewClient creates a Client instance with the given oauth secret used as a access token.
// The given number of retries and a duration of sleep sets how many times the client should invoke the request until
// it gets response with a code < 404 and how long it should sleep between every request attempt
func NewClient(oauthSecret []byte, retries int, sleep time.Duration) *Client {
	token := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: string(oauthSecret)},
	)

	return &Client{
		Client:  gogh.NewClient(oauth2.NewClient(context.Background(), token)),
		Retries: retries,
		Sleep:   sleep,
	}
}

// GetPermissionLevel retrieves the specific permission level a collaborator has for a given repository.
func (c *Client) GetPermissionLevel(owner, repo, user string) (*gogh.RepositoryPermissionLevel, error) {
	var permissionLevel *gogh.RepositoryPermissionLevel
	var err error

	c.do(func() *gogh.Response {
		var response *gogh.Response
		permissionLevel, response, err = c.Client.Repositories.GetPermissionLevel(context.Background(), owner, repo, user)
		return response
	})

	return permissionLevel, err
}

// GetPullRequest retrieves information about a single pull request.
func (c *Client) GetPullRequest(owner, repo string, prNumber int) (*gogh.PullRequest, error) {
	var pr *gogh.PullRequest
	var err error

	c.do(func() *gogh.Response {
		var response *gogh.Response
		pr, response, err = c.Client.PullRequests.Get(context.Background(), owner, repo, prNumber)
		return response
	})

	return pr, err
}

// ListPullRequestFiles lists the changed files in a pull request.
func (c *Client) ListPullRequestFiles(owner, repo string, prNumber int) ([]scm.ChangedFile, error) {
	var files []*gogh.CommitFile
	var err error

	c.do(func() *gogh.Response {
		var response *gogh.Response
		files, response, err = c.Client.PullRequests.ListFiles(context.Background(), owner, repo, prNumber, nil)
		return response
	})

	changedFiles := make([]scm.ChangedFile, 0, len(files))
	for _, file := range files {
		changedFiles = append(changedFiles, scm.ChangedFile{Name: *file.Filename, Status: *file.Status})
	}

	return changedFiles, err
}

// ListIssueComments lists all comments on the specified issue.
func (c *Client) ListIssueComments(issue scm.RepositoryIssue) ([]*gogh.IssueComment, error) {
	var comments []*gogh.IssueComment
	var err error

	c.do(func() *gogh.Response {
		var response *gogh.Response
		comments, response, err =
			c.Client.Issues.ListComments(context.Background(), issue.Owner, issue.RepoName, issue.Number, &gogh.IssueListCommentsOptions{})
		return response
	})

	return comments, err
}

// CreateIssueComment creates a new comment on the specified issue.
func (c *Client) CreateIssueComment(issue scm.RepositoryIssue, commentMsg *string) error {
	var err error
	comment := &gogh.IssueComment{
		Body: commentMsg,
	}

	c.do(func() *gogh.Response {
		var response *gogh.Response
		_, response, err = c.Client.Issues.CreateComment(context.Background(), issue.Owner, issue.RepoName, issue.Number, comment)
		return response
	})

	return err
}

// CreateStatus creates a new status for a repository at the specified reference represented by a RepositoryChange
func (c *Client) CreateStatus(change scm.RepositoryChange, repoStatus *gogh.RepoStatus) error {
	var err error

	c.do(func() *gogh.Response {
		var response *gogh.Response
		_, response, err =
			c.Client.Repositories.CreateStatus(context.Background(), change.Owner, change.RepoName, change.Hash, repoStatus)
		return response
	})

	return err
}

func (c *Client) do(invoker http.RequestInvoker) {
	http.Do(c.Retries, c.Sleep, invoker)
}

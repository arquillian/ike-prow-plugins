package ghclient

import (
	"context"

	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	gogh "github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"fmt"
)

type client struct {
	log log.Logger
	gh  *gogh.Client
}

// Client manages communication with the GitHub API.
type Client interface {
	GetPermissionLevel(owner, repo, user string) (*gogh.RepositoryPermissionLevel, error)
	GetPullRequest(owner, repo string, prNumber int) (*gogh.PullRequest, error)
	ListPullRequestFiles(owner, repo string, prNumber int) ([]scm.ChangedFile, error)
	GetPullRequestReviews(owner, repo string, prNumber int) ([]*gogh.PullRequestReview, error)
	ListIssueComments(issue scm.RepositoryIssue) ([]*gogh.IssueComment, error)
	CreateIssueComment(issue scm.RepositoryIssue, commentMsg *string) error
	CreateStatus(change scm.RepositoryChange, repoStatus *gogh.RepoStatus) error
	ListPullRequestLabels(owner, repo string, prNumber int) ([]*gogh.Label, error)
	AddPullRequestLabel(owner, repo string, prNumber int, label []string) ([]*gogh.Label, error)
	RemovePullRequestLabel(owner, repo string, prNumber int, label string) error
	// This method is intended to be used by client decorators which need access to GH API methods not (yet)
	// exposed by this interface
	unwrap() *gogh.Client
}

// NewOauthClient creates a Client instance with the given oauth secret used as a access token. Underneath
// it creates go-github client which is used as delegate
func NewOauthClient(oauthSecret []byte, log log.Logger) Client {
	token := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: string(oauthSecret)})
	oauthClient := gogh.NewClient(oauth2.NewClient(context.Background(), token))
	return NewClient(oauthClient, log)
}

// NewClient creates a Client instance with the given instance of go-github client which will be used as a delegate
func NewClient(c *gogh.Client, log log.Logger) Client {
	return &client{gh: c, log: log}
}

// GetPermissionLevel retrieves the specific permission level a collaborator has for a given repository.
func (c client) GetPermissionLevel(owner, repo, user string) (*gogh.RepositoryPermissionLevel, error) {
	permissionLevel, response, err := c.gh.Repositories.GetPermissionLevel(context.Background(), owner, repo, user)
	return permissionLevel, c.checkHTTPCode(response, err)
}

// GetPullRequest retrieves information about a single pull request.
func (c client) GetPullRequest(owner, repo string, prNumber int) (*gogh.PullRequest, error) {
	pr, response, err := c.gh.PullRequests.Get(context.Background(), owner, repo, prNumber)
	return pr, c.checkHTTPCode(response, err)
}

// GetPullRequestReviews retrieves a list of reviews submitted to the pull request.
func (c client) GetPullRequestReviews(owner, repo string, prNumber int) ([]*gogh.PullRequestReview, error) {
	prReviews, response, err := c.gh.PullRequests.ListReviews(context.Background(), owner, repo, prNumber, nil)
	return prReviews, c.checkHTTPCode(response, err)
}

// ListPullRequestFiles lists the changed files in a pull request.
func (c client) ListPullRequestFiles(owner, repo string, prNumber int) ([]scm.ChangedFile, error) {
	files, response, err := c.gh.PullRequests.ListFiles(context.Background(), owner, repo, prNumber, nil)
	changedFiles := make([]scm.ChangedFile, 0, len(files))
	for _, file := range files {
		changedFiles = append(changedFiles, *scm.NewChangedFile(file))
	}
	return changedFiles, c.checkHTTPCode(response, err)
}

// ListIssueComments lists all comments on the specified issue.
func (c client) ListIssueComments(issue scm.RepositoryIssue) ([]*gogh.IssueComment, error) {
	comments, response, err :=
		c.gh.Issues.ListComments(context.Background(), issue.Owner, issue.RepoName, issue.Number, &gogh.IssueListCommentsOptions{})

	return comments, c.checkHTTPCode(response, err)
}

// CreateIssueComment creates a new comment on the specified issue.
func (c client) CreateIssueComment(issue scm.RepositoryIssue, commentMsg *string) error {
	comment := &gogh.IssueComment{
		Body: commentMsg,
	}
	_, response, err := c.gh.Issues.CreateComment(context.Background(), issue.Owner, issue.RepoName, issue.Number, comment)
	return c.checkHTTPCode(response, err)
}

// CreateStatus creates a new status for a repository at the specified reference represented by a RepositoryChange
func (c client) CreateStatus(change scm.RepositoryChange, repoStatus *gogh.RepoStatus) error {
	_, response, err :=
		c.gh.Repositories.CreateStatus(context.Background(), change.Owner, change.RepoName, change.Hash, repoStatus)
	return c.checkHTTPCode(response, err)
}

func (c client) ListPullRequestLabels(owner, repo string, prNumber int) ([]*gogh.Label, error) {
	labels, response, err := c.gh.Issues.ListLabelsByIssue(context.Background(), owner, repo, prNumber, nil)
	return labels, c.logHTTPError(response, err)
}

func (c client) AddPullRequestLabel(owner, repo string, prNumber int, label []string) ([]*gogh.Label, error) {
	labels, response, err := c.gh.Issues.AddLabelsToIssue(context.Background(), owner, repo, prNumber, label)
	return labels, c.logHTTPError(response, err)
}

func (c client) RemovePullRequestLabel(owner, repo string, prNumber int, label string) error {
	response, err := c.gh.Issues.RemoveLabelForIssue(context.Background(), owner, repo, prNumber, label)
	return c.logHTTPError(response, err)
}

func (c client) unwrap() *gogh.Client {
	return c.gh
}

func (c client) checkHTTPCode(response *gogh.Response, e error) error {
	if e == nil && response != nil && response.StatusCode >= 404 {
		return fmt.Errorf("server responded with %d status", response.StatusCode)
	}
	return e
}

package ghclient

import (
	"context"

	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	gogh "github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
)

type client struct {
	logger    log.Logger
	gh        *gogh.Client
	allAround aroundFunction
}

// Client manages communication with the GitHub API.
type Client interface {
	GetPermissionLevel(owner, repo, user string) (*gogh.RepositoryPermissionLevel, error)
	GetPullRequest(owner, repo string, prNumber int) (*gogh.PullRequest, error)
	ListPullRequestFiles(owner, repo string, prNumber int) ([]scm.ChangedFile, error)
	GetPullRequestReviews(owner, repo string, prNumber int) ([]*gogh.PullRequestReview, error)
	ListIssueComments(issue scm.RepositoryIssue) ([]*gogh.IssueComment, error)
	CreateIssueComment(issue scm.RepositoryIssue, commentMsg *string) error
	EditIssueComment(issue scm.RepositoryIssue, commentID int64, commentMsg *string) error
	CreateStatus(change scm.RepositoryChange, repoStatus *gogh.RepoStatus) error
	AddPullRequestLabel(change scm.RepositoryChange, prNumber int, label []string) error
	RemovePullRequestLabel(change scm.RepositoryChange, prNumber int, label string) error
	EditPullRequest(*gogh.PullRequest) error
	GetRateLimit() (*gogh.RateLimits, error)

	RegisterAroundFunctions(aroundCreators ...AroundFunctionCreator)
}

// NewOauthClient creates a Client instance with the given oauth secret used as a access token. Underneath
// it creates go-github client which is used as delegate
func NewOauthClient(oauthSecret []byte, logger log.Logger) Client {
	token := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: string(oauthSecret)})
	oauthClient := gogh.NewClient(oauth2.NewClient(context.Background(), token))
	return NewClient(oauthClient, logger)
}

// NewClient creates a Client instance with the given instance of go-github client which will be used as a delegate
func NewClient(c *gogh.Client, logger log.Logger) Client {
	return &client{gh: c, logger: logger, allAround: emptyAround}
}

// AroundFunctionCreator creates function that does operations around nested inner function
type AroundFunctionCreator interface {
	createAroundFunction(f aroundFunction) aroundFunction
}
type aroundFunction func(doFunction doFunction) doFunction
type doFunction func(context aroundContext) (func(), *gogh.Response, error)
type aroundContext struct {
	pageNumber int
}

var emptyAround = func(doFunction doFunction) doFunction {
	return doFunction
}

// RegisterAroundFunctions registers instances of AroundFunctionCreator
func (c *client) RegisterAroundFunctions(aroundCreators ...AroundFunctionCreator) {
	for _, around := range aroundCreators {
		c.allAround = around.createAroundFunction(c.allAround)
	}
}

func (c *client) do(function doFunction) error {
	around := c.allAround(function)
	_, _, e := around(aroundContext{})
	return e
}

// GetPermissionLevel retrieves the specific permission level a collaborator has for a given repository.
func (c *client) GetPermissionLevel(owner, repo, user string) (*gogh.RepositoryPermissionLevel, error) {
	var permissionLevel *gogh.RepositoryPermissionLevel

	err := c.do(func(aroundContext aroundContext) (func(), *gogh.Response, error) {
		level, response, e := c.gh.Repositories.GetPermissionLevel(context.Background(), owner, repo, user)
		return func() {
			permissionLevel = level
		}, response, c.checkHTTPCode(response, e)
	})

	return permissionLevel, err
}

// GetPullRequest retrieves information about a single pull request.
func (c *client) GetPullRequest(owner, repo string, prNumber int) (*gogh.PullRequest, error) {
	var pullRequest *gogh.PullRequest

	err := c.do(func(aroundContext aroundContext) (func(), *gogh.Response, error) {
		pr, response, e := c.gh.PullRequests.Get(context.Background(), owner, repo, prNumber)
		return func() {
			pullRequest = pr
		}, response, c.checkHTTPCode(response, e)
	})

	return pullRequest, err
}

// GetPullRequestReviews retrieves a list of reviews submitted to the pull request.
func (c *client) GetPullRequestReviews(owner, repo string, prNumber int) ([]*gogh.PullRequestReview, error) {
	prReviews := make([]*gogh.PullRequestReview, 0)

	err := c.do(func(aroundCtx aroundContext) (func(), *gogh.Response, error) {
		reviews, response, e := c.gh.PullRequests.ListReviews(context.Background(), owner, repo, prNumber, listOpts(aroundCtx))
		return func() {
			prReviews = append(prReviews, reviews...)
		}, response, c.checkHTTPCode(response, e)
	})

	return prReviews, err
}

// ListPullRequestFiles lists the changed files in a pull request.
func (c *client) ListPullRequestFiles(owner, repo string, prNumber int) ([]scm.ChangedFile, error) {
	changedFiles := make([]scm.ChangedFile, 0)

	err := c.do(func(aroundCtx aroundContext) (func(), *gogh.Response, error) {
		files, response, e := c.gh.PullRequests.ListFiles(context.Background(), owner, repo, prNumber, listOpts(aroundCtx))
		return func() {
			for _, file := range files {
				changedFiles = append(changedFiles, *scm.NewChangedFile(file))
			}
		}, response, c.checkHTTPCode(response, e)
	})

	return changedFiles, err
}

// ListIssueComments lists all comments on the specified issue.
func (c *client) ListIssueComments(issue scm.RepositoryIssue) ([]*gogh.IssueComment, error) {
	allComments := make([]*gogh.IssueComment, 0)

	err := c.do(func(aroundCtx aroundContext) (func(), *gogh.Response, error) {
		commentsOpt := &gogh.IssueListCommentsOptions{ListOptions: *listOpts(aroundCtx)}
		comments, response, e := c.gh.Issues.ListComments(context.Background(), issue.Owner, issue.RepoName, issue.Number, commentsOpt)
		return func() {
			allComments = append(allComments, comments...)
		}, response, c.checkHTTPCode(response, e)
	})

	return allComments, err
}

// CreateIssueComment creates a new comment on the specified issue.
func (c *client) CreateIssueComment(issue scm.RepositoryIssue, commentMsg *string) error {
	comment := &gogh.IssueComment{
		Body: commentMsg,
	}
	err := c.do(func(aroundContext aroundContext) (func(), *gogh.Response, error) {
		_, response, e := c.gh.Issues.CreateComment(context.Background(), issue.Owner, issue.RepoName, issue.Number, comment)
		return func() {}, response, c.checkHTTPCode(response, e)
	})

	return err
}

// EditIssueComment edits an already existing comment in the given issue.
func (c *client) EditIssueComment(issue scm.RepositoryIssue, commentID int64, commentMsg *string) error {
	comment := &gogh.IssueComment{
		Body: commentMsg,
	}
	err := c.do(func(aroundContext aroundContext) (func(), *gogh.Response, error) {
		_, response, e := c.gh.Issues.EditComment(context.Background(), issue.Owner, issue.RepoName, commentID, comment)
		return func() {}, response, c.checkHTTPCode(response, e)
	})

	return err
}

// CreateStatus creates a new status for a repository at the specified reference represented by a RepositoryChange
func (c *client) CreateStatus(change scm.RepositoryChange, repoStatus *gogh.RepoStatus) error {
	err := c.do(func(aroundContext aroundContext) (func(), *gogh.Response, error) {
		_, response, e :=
			c.gh.Repositories.CreateStatus(context.Background(), change.Owner, change.RepoName, change.Hash, repoStatus)
		return func() {}, response, c.checkHTTPCode(response, e)
	})

	return err
}

func (c *client) EditPullRequest(pr *gogh.PullRequest) error {
	err := c.do(func(aroundContext aroundContext) (func(), *gogh.Response, error) {
		_, response, e :=
			c.gh.PullRequests.Edit(context.Background(), *pr.Base.Repo.Owner.Login, *pr.Base.Repo.Name, *pr.Number, pr)
		return func() {}, response, c.checkHTTPCode(response, e)
	})

	return err
}

func (c *client) AddPullRequestLabel(change scm.RepositoryChange, prNumber int, label []string) error {
	err := c.do(func(aroundContext aroundContext) (func(), *gogh.Response, error) {
		_, response, e := c.gh.Issues.AddLabelsToIssue(context.Background(), change.Owner, change.RepoName, prNumber, label)
		return func() {}, response, c.checkHTTPCode(response, e)
	})

	return err
}

func (c *client) RemovePullRequestLabel(change scm.RepositoryChange, prNumber int, label string) error {
	err := c.do(func(aroundContext aroundContext) (func(), *gogh.Response, error) {
		response, e := c.gh.Issues.RemoveLabelForIssue(context.Background(), change.Owner, change.RepoName, prNumber, label)
		return func() {}, response, c.checkHTTPCode(response, e)
	})

	return err
}

// GetRateLimits retrieves the rate limits for the current GH client
func (c *client) GetRateLimit() (*gogh.RateLimits, error) {
	limits, _, err := c.gh.RateLimits(context.Background())
	return limits, err
}

func (c *client) checkHTTPCode(response *gogh.Response, e error) error {
	if e == nil && response != nil && response.StatusCode >= 404 {
		return fmt.Errorf("server responded with %d status", response.StatusCode)
	}
	return e
}

func listOpts(aroundContext aroundContext) *gogh.ListOptions {
	options := &gogh.ListOptions{PerPage: 100}
	options.Page = aroundContext.pageNumber
	return options
}

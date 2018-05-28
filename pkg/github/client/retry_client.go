package ghclient

import (
	"errors"
	"fmt"
	"time"

	"github.com/arquillian/ike-prow-plugins/pkg/retry"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	gogh "github.com/google/go-github/github"
)

type retryClient struct {
	Client
	retries int
	sleep   time.Duration
}

// NewRetryClient creates a client wrapped with retry logic
func NewRetryClient(c Client, retries int, sleep time.Duration) Client {
	return &retryClient{
		Client:  c,
		retries: retries,
		sleep:   sleep,
	}
}

func (r retryClient) retry(toRetry func() error) error {
	errs := retry.Do(r.retries, r.sleep, func() error {
		return toRetry()
	})

	if len(errs) == r.retries {
		msg := fmt.Sprintf("all %d attempts of sending a request failed. See the errors:", r.retries)
		for index, e := range errs {
			msg = msg + fmt.Sprintf("\n%d. [%s]", index+1, e.Error())
		}
		return errors.New(msg)
	}

	return nil
}

func (r retryClient) GetPermissionLevel(owner, repo, user string) (*gogh.RepositoryPermissionLevel, error) {
	var level *gogh.RepositoryPermissionLevel
	err := r.retry(func() error {
		var e error
		level, e = r.Client.GetPermissionLevel(owner, repo, user)
		return e
	})
	return level, err
}

func (r retryClient) GetPullRequest(owner, repo string, prNumber int) (*gogh.PullRequest, error) {
	var pr *gogh.PullRequest
	err := r.retry(func() error {
		var e error
		pr, e = r.Client.GetPullRequest(owner, repo, prNumber)
		return e
	})
	return pr, err
}

func (r retryClient) GetPullRequestReviews(owner, repo string, prNumber int) ([]*gogh.PullRequestReview, error) {
	var reviews []*gogh.PullRequestReview
	err := r.retry(func() error {
		var e error
		reviews, e = r.Client.GetPullRequestReviews(owner, repo, prNumber)
		return e
	})
	return reviews, err
}

func (r retryClient) ListPullRequestFiles(owner, repo string, prNumber int) ([]scm.ChangedFile, error) {
	var files []scm.ChangedFile
	err := r.retry(func() error {
		var e error
		files, e = r.Client.ListPullRequestFiles(owner, repo, prNumber)
		return e
	})
	return files, err
}

func (r retryClient) ListIssueComments(issue scm.RepositoryIssue) ([]*gogh.IssueComment, error) {
	var issues []*gogh.IssueComment
	err := r.retry(func() error {
		var e error
		issues, e = r.Client.ListIssueComments(issue)
		return e
	})
	return issues, err
}

func (r retryClient) CreateIssueComment(issue scm.RepositoryIssue, commentMsg *string) error {
	return r.retry(func() error {
		return r.Client.CreateIssueComment(issue, commentMsg)
	})
}

func (r retryClient) CreateStatus(change scm.RepositoryChange, repoStatus *gogh.RepoStatus) error {
	return r.retry(func() error {
		return r.Client.CreateStatus(change, repoStatus)
	})
}

// GetRateLimits retrieves the rate limits for the current GH client
func (r retryClient) GetRateLimit() (*gogh.RateLimits, error) {
	var limits *gogh.RateLimits
	err := r.retry(func() error {
		var e error
		limits, e = r.Client.GetRateLimit()
		return e
	})
	return limits, err
}

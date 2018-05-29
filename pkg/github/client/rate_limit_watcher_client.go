package ghclient

import (
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	gogh "github.com/google/go-github/github"
)

type rateLimitWatcher struct {
	Client
	log       log.Logger
	threshold int
}

// NewRateLimitWatcherClient wraps github client with calls watching GH API rate limits
func NewRateLimitWatcherClient(c Client, log log.Logger, threshold int) Client {
	return &rateLimitWatcher{Client: c, log: log, threshold: threshold}
}

func (r rateLimitWatcher) logRateLimitsAfter(f func()) {
	f()
	r.logRateLimits()
}

// GetRateLimits retrieves the rate limits for the current GH client
func (r rateLimitWatcher) GetRateLimit() (*gogh.RateLimits, error) {
	return r.Client.GetRateLimit()
}

func (r rateLimitWatcher) logRateLimits() {
	limits, e := r.GetRateLimit()
	if e != nil {
		r.log.Errorf("failed to load rate limits %s", e)
		return
	}
	core := limits.GetCore()
	if core.Remaining < r.threshold {
		r.log.Warnf("reaching limit for GH API calls. %d/%d left. resetting at [%s]", core.Remaining, core.Limit, core.Reset.Format("2006-01-01 15:15:15"))
	}
}

func (r rateLimitWatcher) GetPermissionLevel(owner, repo, user string) (*gogh.RepositoryPermissionLevel, error) {
	var level *gogh.RepositoryPermissionLevel
	var err error
	r.logRateLimitsAfter(func() {
		level, err = r.Client.GetPermissionLevel(owner, repo, user)
	})
	return level, err
}

func (r rateLimitWatcher) GetPullRequest(owner, repo string, prNumber int) (*gogh.PullRequest, error) {
	var pr *gogh.PullRequest
	var err error
	r.logRateLimitsAfter(func() {
		pr, err = r.Client.GetPullRequest(owner, repo, prNumber)
	})
	return pr, err
}

func (r rateLimitWatcher) GetPullRequestReviews(owner, repo string, prNumber int) ([]*gogh.PullRequestReview, error) {
	var reviews []*gogh.PullRequestReview
	var err error
	r.logRateLimitsAfter(func() {
		reviews, err = r.Client.GetPullRequestReviews(owner, repo, prNumber)
	})
	return reviews, err
}

func (r rateLimitWatcher) ListPullRequestFiles(owner, repo string, prNumber int) ([]scm.ChangedFile, error) {
	var files []scm.ChangedFile
	var err error
	r.logRateLimitsAfter(func() {
		files, err = r.Client.ListPullRequestFiles(owner, repo, prNumber)
	})
	return files, err
}

func (r rateLimitWatcher) ListIssueComments(issue scm.RepositoryIssue) ([]*gogh.IssueComment, error) {
	var issues []*gogh.IssueComment
	var err error
	r.logRateLimitsAfter(func() {
		issues, err = r.Client.ListIssueComments(issue)
	})
	return issues, err
}

func (r rateLimitWatcher) CreateIssueComment(issue scm.RepositoryIssue, commentMsg *string) error {
	var err error
	r.logRateLimitsAfter(func() {
		err = r.Client.CreateIssueComment(issue, commentMsg)
	})
	return err
}

func (r rateLimitWatcher) CreateStatus(change scm.RepositoryChange, repoStatus *gogh.RepoStatus) error {
	var err error
	r.logRateLimitsAfter(func() {
		err = r.Client.CreateStatus(change, repoStatus)
	})
	return err
}

func (r rateLimitWatcher) ListPullRequestLabels(change scm.RepositoryChange, prNumber int) ([]*gogh.Label, error) {
	var labels []*gogh.Label
	var err error
	r.logRateLimitsAfter(func() {
		labels, err = r.Client.ListPullRequestLabels(change, prNumber)
	})
	return labels, err
}

func (r rateLimitWatcher) AddPullRequestLabel(change scm.RepositoryChange, prNumber int, label []string) ([]*gogh.Label, error) {
	var labels []*gogh.Label
	var err error
	r.logRateLimitsAfter(func() {
		labels, err = r.Client.AddPullRequestLabel(change, prNumber, label)
	})
	return labels, err
}

func (r rateLimitWatcher) RemovePullRequestLabel(change scm.RepositoryChange, prNumber int, label string) error {
	var err error
	r.logRateLimitsAfter(func() {
		err = r.Client.RemovePullRequestLabel(change, prNumber, label)
	})
	return err
}

package ghclient

import (
	"context"

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

func (r rateLimitWatcher) logRateLimits() {
	ghclient := r.unwrap()
	limits, _, e := ghclient.RateLimits(context.Background())
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

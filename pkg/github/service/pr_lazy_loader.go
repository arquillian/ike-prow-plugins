package ghservice

import (
	ghclient "github.com/arquillian/ike-prow-plugins/pkg/github/client"
	gogh "github.com/google/go-github/github"
)

// PullRequestLazyLoader represents a lazy loader of pull request information - is loaded when needed and only once
type PullRequestLazyLoader struct {
	Client ghclient.Client
	RepoOwner,
	RepoName string
	Number      int
	pullRequest *gogh.PullRequest
	err         error
}

// NewPullRequestLazyLoaderFromComment creates a new instance of PullRequestLazyLoader with information retrieved from the given IssueCommentEvent
func NewPullRequestLazyLoaderFromComment(client ghclient.Client, comment *gogh.IssueCommentEvent) *PullRequestLazyLoader {
	return &PullRequestLazyLoader{
		Client:    client,
		RepoOwner: *comment.Repo.Owner.Login,
		RepoName:  *comment.Repo.Name,
		Number:    *comment.Issue.Number,
	}
}

// NewPullRequestLazyLoaderWithPR creates a new instance of PullRequestLazyLoader with the given already loaded gogh.PullRequest instance
func NewPullRequestLazyLoaderWithPR(client ghclient.Client, pullRequest *gogh.PullRequest) *PullRequestLazyLoader {
	return &PullRequestLazyLoader{
		Client:      client,
		RepoOwner:   *pullRequest.Base.Repo.Owner.Login,
		RepoName:    *pullRequest.Base.Repo.Name,
		Number:      *pullRequest.Number,
		pullRequest: pullRequest,
	}
}

// Load loads information about pull request - if not already retrieved from GH then it gets it and stores, then it uses
// this stored instance
func (r *PullRequestLazyLoader) Load() (*gogh.PullRequest, error) {
	if r.pullRequest == nil {
		r.pullRequest, r.err = r.Client.GetPullRequest(r.RepoOwner, r.RepoName, r.Number)
	}
	return r.pullRequest, r.err
}

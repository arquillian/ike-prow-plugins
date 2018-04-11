package github

import (
	gogh "github.com/google/go-github/github"
)

// PullRequestLazyLoader represents a lazy loader of pull request information - is loaded when needed and only once
type PullRequestLazyLoader struct {
	Client *Client
	RepoOwner,
	RepoName string
	Number      int
	pullRequest *gogh.PullRequest
	err         error
}

// NewPullRequestLazyLoader creates a new instance of PullRequestLazyLoader with information retrieved from the given IssueCommentEvent
func NewPullRequestLazyLoader(client *Client, comment *gogh.IssueCommentEvent) *PullRequestLazyLoader {
	return &PullRequestLazyLoader{
		Client:    client,
		RepoOwner: *comment.Repo.Owner.Login,
		RepoName:  *comment.Repo.Name,
		Number:    *comment.Issue.Number,
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

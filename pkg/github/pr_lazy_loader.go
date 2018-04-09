package github

import (
	gogh "github.com/google/go-github/github"
)

// PullRequestLoader represents a lazy loader of pull request information - is loaded when needed and only once
type PullRequestLoader struct {
	Client      *Client
	RepoOwner   string
	RepoName    string
	Number      int
	pullRequest *gogh.PullRequest
	err         error
}

// NewPullRequestLoader creates a new instance of PullRequestLoader with information retrieved from the given IssueCommentEvent
func NewPullRequestLoader(client *Client, prComment *gogh.IssueCommentEvent) *PullRequestLoader {
	return &PullRequestLoader{
		Client:    client,
		RepoOwner: *prComment.Repo.Owner.Login,
		RepoName:  *prComment.Repo.Name,
		Number:    *prComment.Issue.Number,
	}
}

// Load loads information about pull request - if not already retrieved from GH then it gets it and stores, then it uses
// this stored instance
func (r *PullRequestLoader) Load() (*gogh.PullRequest, error) {
	if r.pullRequest == nil {
		r.pullRequest, r.err = r.Client.GetPullRequest(r.RepoOwner, r.RepoName, r.Number)
	}
	return r.pullRequest, r.err
}

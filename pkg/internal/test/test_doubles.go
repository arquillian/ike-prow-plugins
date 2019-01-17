package test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	ghclient "github.com/arquillian/ike-prow-plugins/pkg/github/client"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	gogh "github.com/google/go-github/github"
	"github.com/onsi/ginkgo"
)

// This package is intended to keep helper functions used across the tests. Shouldn't be used for production code

// nolint
func LoadFromFile(filePath string) []byte {
	payload, err := ioutil.ReadFile(filePath)
	if err != nil {
		ginkgo.Fail(fmt.Sprintf("Unable to load test fixture. Reason: %q", err))
	}
	return payload
}

// nolint
func LoadedFrom(filePath string) string {
	return string(LoadFromFile(filePath))
}

// nolint
func FromFile(filePath string) io.Reader {
	file, err := os.Open(filePath)
	if err != nil {
		ginkgo.Fail(fmt.Sprintf("Unable to load test fixture. Reason: %q", err))
	}
	return file
}

// NewPullRequest creates an instance of gogh.PullRequest with the given values
func NewPullRequest(repoOwner, repoName, sha, creator string) *gogh.PullRequest {
	return &gogh.PullRequest{
		Base: &gogh.PullRequestBranch{
			Repo: &gogh.Repository{
				Name: utils.String(repoName),
				Owner: &gogh.User{
					Login: utils.String(repoOwner),
				},
			},
		},
		Head: &gogh.PullRequestBranch{
			SHA: utils.String(sha),
		},
		User: &gogh.User{
			Login: utils.String(creator),
		},
	}
}

// LoadIssueCommentEvent creates an instance of gogh.PullRequestEvent with the given values
func LoadIssueCommentEvent(filePath string) *gogh.IssueCommentEvent {
	var event gogh.IssueCommentEvent
	payload := LoadFromFile(filePath)
	if err := json.Unmarshal(payload, &event); err != nil {
		ginkgo.Fail(fmt.Sprintf("Failed while parsing '%q' event with payload: %+v cause %q.", github.IssueComment, event, err))
	}
	return &event
}

// LoadPullRequestEvent creates an instance of gogh.PullRequestEvent with the given values
func LoadPullRequestEvent(filePath string) *gogh.PullRequestEvent {
	var event gogh.PullRequestEvent
	payload := LoadFromFile(filePath)
	if err := json.Unmarshal(payload, &event); err != nil {
		ginkgo.Fail(fmt.Sprintf("Failed while parsing '%q' event with payload: %+v cause %q.", github.PullRequest, event, err))
	}
	return &event
}

// NewDefaultGitHubClient creates a GH client with default go-github client (without any authentication token)
func NewDefaultGitHubClient() ghclient.Client {
	client := ghclient.NewClient(gogh.NewClient(nil), log.NewTestLogger())
	client.RegisterAroundFunctions(ghclient.NewPaginationChecker())
	return client
}

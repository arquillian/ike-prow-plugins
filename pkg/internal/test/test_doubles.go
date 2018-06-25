package test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
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
func FromFile(filePath string) io.Reader {
	file, err := os.Open(filePath)
	if err != nil {
		ginkgo.Fail(fmt.Sprintf("Unable to load test fixture. Reason: %q", err))
	}
	return file
}

// LoadIssueCommentEvent creates an instance of gogh.PullRequestEvent with the given values
func LoadIssueCommentEvent(filePath string) *gogh.IssueCommentEvent {
	var event gogh.IssueCommentEvent
	payload := LoadFromFile(filePath)
	if err := json.Unmarshal(payload, &event); err != nil {
		ginkgo.Fail(fmt.Sprintf("Failed while parsing '%q' event with payload: %q cause %q.", github.IssueComment, event, err))
	}
	return &event
}

// LoadPullRequestEvent creates an instance of gogh.PullRequestEvent with the given values
func LoadPullRequestEvent(filePath string) *gogh.PullRequestEvent {
	var event gogh.PullRequestEvent
	payload := LoadFromFile(filePath)
	if err := json.Unmarshal(payload, &event); err != nil {
		ginkgo.Fail(fmt.Sprintf("Failed while parsing '%q' event with payload: %q cause %q.", github.PullRequest, event, err))
	}
	return &event
}

// NewDefaultGitHubClient creates a GH client with default go-github client (without any authentication token)
func NewDefaultGitHubClient() ghclient.Client {
	client := ghclient.NewClient(gogh.NewClient(nil), log.NewTestLogger())
	client.RegisterAroundFunctions(ghclient.NewPaginationChecker())
	return client
}

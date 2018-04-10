package test

import (
	"net/http"
	"strings"

	"github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

// EnsureGockRequestsHaveBeenMatched checks if all requests have been matched and all mocked requests have been used in the test
func EnsureGockRequestsHaveBeenMatched() {
	gomega.Expect(gock.GetUnmatchedRequests()).To(gomega.BeEmpty(), "Have no unmatched requests")
}

// NonExistingRawGitHubFiles mocks any matching path suffix when calling "https://raw.githubusercontent.com" with 404 response
func NonExistingRawGitHubFiles(pathSuffixes ...string) {
	for _, pathSuffix := range pathSuffixes {
		gock.New("https://raw.githubusercontent.com").
			SetMatcher(fileRequested(pathSuffix)).
			Reply(404)
	}
}

func fileRequested(pathSuffix string) gock.Matcher {
	matcher := gock.NewBasicMatcher()
	matcher.Add(func(req *http.Request, _ *gock.Request) (bool, error) {
		return strings.HasSuffix(req.URL.Path, pathSuffix), nil
	})
	return matcher
}

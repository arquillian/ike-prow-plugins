package test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	gogh "github.com/google/go-github/github"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"gopkg.in/h2non/gock.v1"
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

// nolint
func ExpectPayload(payloadAssert ...func(payload map[string]interface{}) bool) gock.Matcher {
	matcher := gock.NewBasicMatcher()
	matcher.Add(func(req *http.Request, _ *gock.Request) (bool, error) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return false, err
		}
		var payload map[string]interface{}
		err = json.Unmarshal(body, &payload)
		var payloadExpectations bool
		for _, assertion := range payloadAssert {
			payloadExpectations = assertion(payload)
			if !payloadExpectations {
				break
			}
		}
		return payloadExpectations, err
	})
	return matcher
}

// NewDiscardOutLogger creates a logger instance not logging any output to Out Writer
func NewDiscardOutLogger() log.Logger {
	nullLogger := logrus.New()
	nullLogger.Out = ioutil.Discard // TODO rethink if we want to discard logging entirely
	return logrus.NewEntry(nullLogger)
}

// NewDefaultGitHubClient creates a GH client with default go-github client (without any authentication token),
// with number of retries set to 3 and sleep duration set to 1 second
func NewDefaultGitHubClient() *github.Client {
	return &github.Client{
		Client:  gogh.NewClient(nil), // TODO with hoverfly/go-vcr we might want to use tokens instead to capture real traffic
		Retries: 3,
		Sleep:   0,
	}
}

// ToHaveBodyContaining asserts if the payload contains the given string
func ToHaveBodyContaining(toContains string) func(statusPayload map[string]interface{}) bool {
	return func(statusPayload map[string]interface{}) bool {
		return gomega.Expect(statusPayload).To(gomega.SatisfyAll(HaveBodyThatContains(toContains)))
	}
}

package test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	gogh "github.com/google/go-github/github"
	"github.com/onsi/ginkgo"
	"github.com/sirupsen/logrus"
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

// NewDiscardOutLogger creates a logger instance not logging any output to Out Writer
func NewDiscardOutLogger() log.Logger {
	nullLogger := logrus.New()
	nullLogger.Out = ioutil.Discard // TODO rethink if we want to discard logging entirely
	return logrus.NewEntry(nullLogger)
}

// NewDefaultGitHubClient creates a GH client with default go-github client (without any authentication token),
// with number of retries set to 3 and no sleep duration as we use it for mocked tests
func NewDefaultGitHubClient() *github.Client {
	return &github.Client{
		Client:  gogh.NewClient(nil), // TODO with hoverfly/go-vcr we might want to use tokens instead to capture real traffic
		Retries: 3,
		Sleep:   0,
	}
}

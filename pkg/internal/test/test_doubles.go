package test

import (
	"fmt"
	"os"
	"gopkg.in/h2non/gock.v1"
	"net/http"
	"io"
	"io/ioutil"
	"github.com/onsi/ginkgo"
	"encoding/json"
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

// nolint
func ExpectStatusCall(payloadAssert func(statusPayload map[string]interface{}) (bool)) gock.Matcher {
	matcher := gock.NewBasicMatcher()
	matcher.Add(func(req *http.Request, _ *gock.Request) (bool, error) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return false, err
		}
		var statusPayload map[string]interface{}
		err = json.Unmarshal(body, &statusPayload)
		payloadExpectations := payloadAssert(statusPayload)
		return payloadExpectations, err
	})
	return matcher
}

// nolint
func CreateNullLogger() *logrus.Entry {
	nullLogger := logrus.New()
	nullLogger.Out =  ioutil.Discard // TODO rethink if we want to discard logging entirely
	return logrus.NewEntry(nullLogger)
}
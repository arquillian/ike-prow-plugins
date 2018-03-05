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
)
// nolint
func EventPayload(payloadFile string) []byte {
	payload, err := ioutil.ReadFile(payloadFile)
	if err != nil {
		ginkgo.Fail(fmt.Sprintf("Unable to load test fixture. Reason: %q", err))
	}
	return payload
}

// nolint
func FromJson(filePath string) io.Reader {
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
package test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

// ExpectPayload creates a gock.Matcher with the given SoftMatchers that asserts a retrieved payload
func ExpectPayload(matchers ...SoftMatcher) gock.Matcher {
	return createPayloadMatcher(matchers)
}

// To softly satisfies all the given matchers
func To(matchers ...SoftMatcher) SoftMatcher {
	return SoftlySatisfyAll(matchers...)
}

func createPayloadMatcher(matchers []SoftMatcher) gock.Matcher {
	matcher := gock.NewBasicMatcher()
	matcher.Add(func(req *http.Request, _ *gock.Request) (bool, error) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return false, err
		}
		var payload map[string]interface{}
		err = json.Unmarshal(body, &payload)
		payloadExpectations := createPayloadAssert(matchers)(payload)
		return payloadExpectations, err
	})
	return matcher
}

func createPayloadAssert(matchers []SoftMatcher) func(statusPayload map[string]interface{}) bool {
	return func(statusPayload map[string]interface{}) bool {
		return gomega.Expect(statusPayload).To(SoftlySatisfyAll(matchers...))
	}
}

// HaveState gets "state" key from map[string]interface{} and compares its value with expectedState
// This matcher is used to verify status update sent to GitHub API
func HaveState(expectedState string) SoftMatcher {
	return TransformWithName(
		func(s map[string]interface{}) interface{} { return s["state"] },
		gomega.Equal(expectedState),
		"state")
}

// HaveDescription gets "description" key from map[string]interface{} and compares its value with expectedReason
// This matcher is used to verify status description sent to GitHub API
func HaveDescription(expectedReason string) SoftMatcher {
	return TransformWithName(
		func(s map[string]interface{}) interface{} { return s["description"] },
		gomega.Equal(expectedReason),
		"description")
}

// HaveContext gets "context" key from map[string]interface{} and compares its value with expectedContext
// This matcher is used to verify status context sent to GitHub API
func HaveContext(expectedContext string) SoftMatcher {
	return TransformWithName(
		func(s map[string]interface{}) interface{} { return s["context"] },
		gomega.Equal(expectedContext),
		"context")
}

// HaveTargetURL gets "target_url" key from map[string]interface{} and compares its value with expectedTargetURL
// This matcher is used to verify status target URL sent to GitHub API
func HaveTargetURL(expectedTargetURL string) SoftMatcher {
	return TransformWithName(
		func(s map[string]interface{}) interface{} { return s["target_url"] },
		gomega.Equal(expectedTargetURL),
		"target_url")
}

// HaveLabels gets "label" key from map[string]interface{} and compares its value with expectedLabels
// This matcher is used to verify status labels sent to GitHub API
func HaveLabels(expectedLabels string) SoftMatcher {
	return TransformWithName(
		func(s map[string]interface{}) interface{} { return s["labels"] },
		gomega.Equal(expectedLabels),
		"labels")
}

// HaveBody gets "body" key from map[string]interface{} and compares its value with expectedBody
// This matcher is used to verify body content sent in request to GitHub API
func HaveBody(expectedBody string) SoftMatcher {
	return TransformWithName(
		func(s map[string]interface{}) interface{} { return s["body"] },
		gomega.Equal(expectedBody),
		"body")
}

// HaveBodyThatContains gets "body" key from map[string]interface{} and checks if its value contains the given string
// This matcher is used to verify body content sent in request to GitHub API
func HaveBodyThatContains(content string) SoftMatcher {
	return TransformWithName(
		func(s map[string]interface{}) interface{} { return s["body"] },
		gomega.ContainSubstring(content),
		"body")
}

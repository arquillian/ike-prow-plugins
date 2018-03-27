package test

import (
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

// HaveState gets "state" key from map[string]interface{} and compares its value with expectedState
// This matcher is used to verify status update sent to GitHub API
func HaveState(expectedState string) types.GomegaMatcher {
	return gomega.WithTransform(func(s map[string]interface{}) interface{} { return s["state"] }, gomega.Equal(expectedState))
}

// HaveDescription gets "description" key from map[string]interface{} and compares its value with expectedReason
// This matcher is used to verify status description sent to GitHub API
func HaveDescription(expectedReason string) types.GomegaMatcher {
	return gomega.WithTransform(func(s map[string]interface{}) interface{} { return s["description"] }, gomega.Equal(expectedReason))
}

// HaveContext gets "context" key from map[string]interface{} and compares its value with expectedContext
// This matcher is used to verify status context sent to GitHub API
func HaveContext(expectedContext string) types.GomegaMatcher {
	return gomega.WithTransform(func(s map[string]interface{}) interface{} { return s["context"] }, gomega.Equal(expectedContext))
}

// HaveBody gets "body" key from map[string]interface{} and compares its value with expectedBody
// This matcher is used to verify body content sent in request to GitHub API
func HaveBody(expectedBody string) types.GomegaMatcher {
	return gomega.WithTransform(func(s map[string]interface{}) interface{} { return s["body"] }, gomega.Equal(expectedBody))
}

// HaveBodyThatContains gets "body" key from map[string]interface{} and checks if its value contains the given string
// This matcher is used to verify body content sent in request to GitHub API
func HaveBodyThatContains(content string) types.GomegaMatcher {
	return gomega.WithTransform(func(s map[string]interface{}) interface{} { return s["body"] }, gomega.ContainSubstring(content))
}

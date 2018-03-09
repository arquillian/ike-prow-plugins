package test

import (
	"github.com/onsi/gomega/types"
	"github.com/onsi/gomega"
)

// HaveState gets "state" key from map[string]interface{} and compares its value with expectedState
// This matcher is used to verify status update sent to GitHub API
func HaveState(expectedState string) types.GomegaMatcher {
	return gomega.WithTransform(func(s map[string]interface{}) interface{} { return s["state"]}, gomega.Equal(expectedState))
}

// HaveDescription gets "description" key from map[string]interface{} and compares its value with expectedState
// This matcher is used to verify status update sent to GitHub API
func HaveDescription(expectedReason string) types.GomegaMatcher {
	return gomega.WithTransform(func(s map[string]interface{}) interface{} { return s["description"]}, gomega.Equal(expectedReason))
}

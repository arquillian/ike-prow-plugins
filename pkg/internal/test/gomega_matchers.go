package test

import (
	"github.com/onsi/gomega/types"
	"github.com/onsi/gomega"
)

func HaveState(expectedState string) types.GomegaMatcher {
	return gomega.WithTransform(func(s map[string]interface{}) interface{} { return s["state"]}, gomega.Equal(expectedState))
}

func HaveDescription(expectedReason string) types.GomegaMatcher {
	return gomega.WithTransform(func(s map[string]interface{}) interface{} { return s["description"]}, gomega.Equal(expectedReason))
}

package wip

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuiteWorkInProgressPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Work-in-progress Plugin Test Suite")
}

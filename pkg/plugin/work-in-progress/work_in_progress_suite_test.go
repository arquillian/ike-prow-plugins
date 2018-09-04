package wip

import (
	"testing"

	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSuiteWorkInProgressPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "Work-in-progress Plugin Test Suite")
}

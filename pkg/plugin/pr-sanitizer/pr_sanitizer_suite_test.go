package prsanitizer

import (
	"testing"

	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSuitePRSanitizerPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "PR-Sanitizer Plugin Test Suite")
}

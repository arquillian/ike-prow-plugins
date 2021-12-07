package prsanitizer

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuitePRSanitizerPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PR-Sanitizer Plugin Test Suite")
}

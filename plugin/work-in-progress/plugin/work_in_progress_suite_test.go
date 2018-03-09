package plugin_test

import (
	"testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSuiteWorkInProgressPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Prow Event Handler Work-in-progress Plugin Suite")
}

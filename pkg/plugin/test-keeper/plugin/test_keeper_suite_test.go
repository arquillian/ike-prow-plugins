package plugin_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSuiteTestKeeperPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Prow Event Handler Test Keeper Plugin Suite")
}

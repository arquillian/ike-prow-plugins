package plugin_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSuiteTestKeeperPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Keeper Prow Plugin Suite")
}

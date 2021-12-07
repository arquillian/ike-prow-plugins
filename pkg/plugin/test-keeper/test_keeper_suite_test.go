package testkeeper_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuiteTestKeeperPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Keeper Prow Plugin Test Suite")
}

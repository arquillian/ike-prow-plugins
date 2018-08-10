package testkeeper_test

import (
	"testing"

	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSuiteTestKeeperPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecWithJUnitReporter(t, "Test Keeper Prow Plugin Test Suite")
}

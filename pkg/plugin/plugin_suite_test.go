package plugin_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSuitePlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ike Prow Plugins Suite")
}

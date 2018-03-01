package plugin_test

import (
	"testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTestKeeper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Prow Event Handler Test Keeper Plugin Suite")
	// TODO configure .gitignore or store in different folder
	//junitReporter := reporters.NewJUnitReporter("junit.xml")
	//RunSpecsWithDefaultAndCustomReporters(t, "Build Tool Detectors Suite", []Reporter{junitReporter})
}

package probeshandler_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuiteProbesHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Probes Handler Suite")
}

const (
	probesEndpoint = "/version"
	defaultVersion = "xxxxxxxx-xxxxxxxxxx"
)

var versionEnv string

var _ = BeforeSuite(func() {
	var found bool
	if versionEnv, found = os.LookupEnv("VERSION"); !found {
		_ = os.Setenv("VERSION", defaultVersion)
	}
})

var _ = AfterSuite(func() {
	_ = os.Setenv("VERSION", versionEnv)
})

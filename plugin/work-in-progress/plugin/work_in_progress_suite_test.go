package plugin_test

import (
	"testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

func TestSuiteWorkInProgressPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Prow Event Handler Work-in-progress Plugin Suite")
}

var Logger *logrus.Entry

var _ = BeforeSuite(func() {
	nullLogger := logrus.New()
	nullLogger.Out = ioutil.Discard
	Logger = logrus.NewEntry(nullLogger)
})

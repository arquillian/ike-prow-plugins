package plugin_test

import (
	"testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

func TestTestKeeper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Prow Test Keeper Plugin Suite")
}

var Logger *logrus.Entry

var _ = BeforeSuite(func() {
	nullLogger := logrus.New()
	nullLogger.Out = ioutil.Discard // TODO rethink if we want to discard logging entirely
	Logger = logrus.NewEntry(nullLogger)
})

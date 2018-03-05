package plugin_test

import (
	"testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTestKeeper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Prow Event Handler WIP Plugin Suite")
}

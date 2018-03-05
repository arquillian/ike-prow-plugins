package plugin_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo/extensions/table"
	"github.com/arquillian/ike-prow-plugins/plugin/work-in-progress/plugin"
	"gopkg.in/h2non/gock.v1"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"github.com/google/go-github/github"
)

var _ = Describe("Test Keeper Plugin features", func() {

	Context("Title verifier", func() {

		var handler *plugin.GitHubWIPPRHandler

		BeforeEach(func() {
			defer gock.Off()

			nullLogger := logrus.New()
			nullLogger.Out = ioutil.Discard
			logger := logrus.NewEntry(nullLogger)

			client := github.NewClient(nil) // TODO with hoverfly/go-vcr we might want to use tokens instead to capture real traffic
			handler = &plugin.GitHubWIPPRHandler{
				Client: client,
				Log:    logger,
			}
		})

		DescribeTable("should recognize PR as work-in-progress if title starts with WIP",
			func(title string) {
				Expect(handler.IsWorkInProgress(&title)).To(BeTrue())
			},
			Entry("Uppercase WIP prefix", "WIP fix(#1): off-by one bug"),
			Entry("lowercase WIP prefix", "wip fix(#1): off-by one bug"),
			Entry("Wip prefix", "Wip fix(#1): off-by one bug"),
		)

		DescribeTable("should not recognize PR as work-in-progress if title doesn't start with WIP",
			func(title string) {
				Expect(handler.IsWorkInProgress(&title)).To(BeFalse())
			},
			Entry("regular PR title", "fix(#1): off-by one bug"),
			Entry("wip", "wip-fix off-by one bug"),
			Entry("empty title", ""),
			Entry("nil title", nil),
		)

		It("should not recognize PR as work-in-progress if title doesn't start with WIP", func() {

		})

	})

})



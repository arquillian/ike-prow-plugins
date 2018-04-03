package plugin_test

import (
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	wip "github.com/arquillian/ike-prow-plugins/pkg/plugin/work-in-progress/plugin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("Work-in-progress Plugin features", func() {

	Context("Title verifier", func() {

		var handler *wip.GitHubWIPPRHandler

		BeforeEach(func() {
			defer gock.Off()

			handler = &wip.GitHubWIPPRHandler{Client: NewDefaultGitHubClient(), BotName: "alien-ike"}
		})

		DescribeTable("should recognize PR as work-in-progress if title starts with WIP",
			func(title string) {
				Expect(handler.IsWorkInProgress(&title)).To(BeTrue())
				Expect(handler.BotName).To(Equal("alien-ike"))
			},
			Entry("Uppercase WIP prefix", "WIP fix(#1): off-by one bug"),
			Entry("Lowercase WIP prefix", "wip fix(#1): off-by one bug"),
			Entry("Wip prefix", "Wip fix(#1): off-by one bug"),
		)

		DescribeTable("should not recognize PR as work-in-progress if title doesn't start with WIP",
			func(title string) {
				Expect(handler.IsWorkInProgress(&title)).To(BeFalse())
				Expect(handler.BotName).To(Equal("alien-ike"))
			},
			Entry("regular PR title", "fix(#1): off-by one bug"),
			Entry("not a supported wip prefix", "wip-fix off-by one bug"),
			Entry("empty title", ""),
			Entry("nil title", nil),
		)

	})

})

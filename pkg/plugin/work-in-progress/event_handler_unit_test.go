package wip_test

import (
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/work-in-progress"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Work-in-progress Plugin features", func() {

	Context("Title verifier", func() {

		DescribeTable("should recognize PR as work-in-progress if title starts with configured or default prefix",
			func(title, expectedPrefix string) {
				prefix, state := wip.GetWorkInProgressPrefix(title, wip.PluginConfiguration{Prefix: []string{`On Hold`}, Combine: true})
				Expect(state).To(BeTrue())
				Expect(prefix).To(Equal(expectedPrefix))
			},
			Entry("Default Wip prefix", "Wip fix(#1): off-by one bug", "Wip"),
			Entry("Custom Work In Progress prefix", "On Hold fix(#1): off-by one bug", "On Hold"),
			Entry("Custom Work In Progress with brackets prefix", "[On Hold] fix(#1): off-by one bug", "[On Hold]"),
		)

		DescribeTable("should recognize PR as work-in-progress if title starts with any default prefix",
			func(title, expectedPrefix string) {
				prefix, state := wip.GetWorkInProgressPrefix(title, wip.PluginConfiguration{})
				Expect(state).To(BeTrue())
				Expect(prefix).To(Equal(expectedPrefix))
			},
			Entry("Uppercase WIP prefix", "WIP fix(#1): off-by one bug", "WIP"),
			Entry("Lowercase WIP prefix", "wip fix(#1): off-by one bug", "wip"),
			Entry("With stars WIP prefix", "*WIP* fix(#1): off-by one bug", "*WIP*"),
			Entry("With multiple stars WIP prefix", "**WIP** fix(#1): off-by one bug", "**WIP**"),
			Entry("Similar to xml comment WIP prefix", "<!--wip--!> fix(#1): off-by one bug", "<!--wip--!>"),
			Entry("With arrows WIP prefix", "--> WIP <-- fix(#1): off-by one bug", "--> WIP <--"),
			Entry("With underscore WIP prefix", "_wip_ fix(#1): off-by one bug", "_wip_"),

			Entry("Wip prefix", "Wip fix(#1): off-by one bug", "Wip"),
			Entry("Wip prefix within square brackets", "[WIP] fix(#1): off-by one bug", "[WIP]"),
			Entry("Wip prefix within brackets", "(WIP) fix(#1): off-by one bug", "(WIP)"),
			Entry("Wip prefix with brackets and colon", "(WIP): fix(#1): off-by one bug", "(WIP):"),
			Entry("Wip prefix with colon ", "WIP: fix(#1): off-by one bug", "WIP:"),
			Entry("Do Not Merge prefix", "Do Not Merge fix(#1): off-by one bug", "Do Not Merge"),
			Entry("Work In Progress prefix", "WORK-IN-PROGRESS fix(#1): off-by one bug", "WORK-IN-PROGRESS"),
		)

		DescribeTable("should not recognize PR as work-in-progress if title doesn't start with any default prefix",
			func(title, expectedPrefix string) {
				prefix, state := wip.GetWorkInProgressPrefix(title, wip.PluginConfiguration{})
				Expect(state).To(BeFalse())
				Expect(prefix).To(Equal(expectedPrefix))
			},
			Entry("regular PR title", "fix(#1): off-by one bug", ""),
			Entry("not a supported wip prefix", "wip-fix off-by one bug", ""),
			Entry("empty title", "", ""),
			Entry("nil title", nil, ""),
		)

		DescribeTable("should not recognize PR as work-in-progress if title starts with default prefix when custom prefix is configured with combine false",
			func(title, expectedPrefix string) {
				prefix, state := wip.GetWorkInProgressPrefix(title, wip.PluginConfiguration{Prefix: []string{`On Hold`}, Combine: false})
				Expect(state).To(BeFalse())
				Expect(prefix).To(Equal(expectedPrefix))
			},
			Entry("Default Wip prefix", "Wip fix(#1): off-by one bug", ""),
		)

	})

})

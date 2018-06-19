package prsanitizer_test

import (
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	prsanitizer "github.com/arquillian/ike-prow-plugins/pkg/plugin/pr-sanitizer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("PR-Sanitizer Plugin features", func() {

	Context("Title verifier", func() {

		var handler *prsanitizer.GitHubLabelsEventsHandler

		BeforeEach(func() {
			handler = &prsanitizer.GitHubLabelsEventsHandler{Client: NewDefaultGitHubClient()}
		})

		DescribeTable("should recognize PR as compliant with semantic commit message if title starts with any default prefix",
			func(title string) {
				Expect(handler.HasTitleWithValidType(prsanitizer.PluginConfiguration{}, title)).To(BeTrue())
			},
			Entry("chore prefix", "chore(#1): add Oyster build script"),
			Entry("docs prefix", "docs(#1): explain hat wobble"),
			Entry("feat prefix", "feat(#1): add beta sequence"),
			Entry("fix prefix", "fix(#1): remove broken confirmation message"),
			Entry("refactor prefix", "refactor(#1): share logic between 4d3d3d3 and flarhgunnstow"),
			Entry("style prefix", "style(#1): convert tabs to spaces"),
			Entry("tests prefix", "test(#1): ensure Tayne retains clothing"),
		)

		DescribeTable("should not recognize PR as compliant with semantic commit message if title doesn't start with any default prefix",
			func(title string) {
				Expect(handler.HasTitleWithValidType(prsanitizer.PluginConfiguration{}, title)).To(BeFalse())
			},
			Entry("not a supported semantic commit prefix", "wip-fix off-by one bug"),
			Entry("empty title", ""),
			Entry("nil title", nil),
		)
	})

})

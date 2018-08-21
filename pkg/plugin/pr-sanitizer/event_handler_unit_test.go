package prsanitizer_test

import (
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/pr-sanitizer"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	"github.com/google/go-github/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("PR-Sanitizer Plugin features", func() {

	Context("Title verifier", func() {

		DescribeTable("should recognize PR as compliant with semantic commit message if title starts with any default prefix",
			func(title string) {
				prefixes := prsanitizer.GetValidTitlePrefixes(prsanitizer.PluginConfiguration{})
				Expect(prsanitizer.HasTitleWithValidType(prefixes, title)).To(BeTrue())
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
				prefixes := prsanitizer.GetValidTitlePrefixes(prsanitizer.PluginConfiguration{})
				Expect(prsanitizer.HasTitleWithValidType(prefixes, title)).To(BeFalse())
			},
			Entry("not a supported semantic commit prefix", "wip-fix off-by one bug"),
			Entry("title starting by fixes", "fixes failing test"),
			Entry("empty title", ""),
			Entry("nil title", nil),
		)
	})

	Context("Description verifier", func() {

		DescribeTable("should recognize issue link presence",
			func(desc string) {
				pr := &github.PullRequest{Body: utils.String(desc)}
				msg := prsanitizer.CheckIssueLinkPresence(pr, prsanitizer.PluginConfiguration{}, log.NewTestLogger())
				Expect(msg).To(BeEmpty())
			},
			Entry("fixes keyword with surrounded text", "PR\r\n\r\nfixes: #1 issue"),
			Entry("Fixes keyword with capital F", "PR description\r\n\r\nFixes: #1"),
			Entry("closed keyword inline", "PR closed: org/repo#1 issue"),
			Entry("resolve keyword with other repo path", "PR\r\nresolve: org/my-repo#1"),
			Entry("only fixed keyword with link", "fixed :  #1"),
			Entry("CLOSES uppercase", "CLOSES: #1"),
			Entry("closes keyword with other repo path inline", "PR closes org/repo#1 issue"),
			Entry("two links", "pr resolves org/my-repo#1 and fixes: #1"),
		)

		DescribeTable("should NOT recognize issue link presence",
			func(desc string) {
				pr := &github.PullRequest{Body: utils.String(desc)}
				msg := prsanitizer.CheckIssueLinkPresence(pr, prsanitizer.PluginConfiguration{}, log.NewTestLogger())
				Expect(msg).NotTo(BeEmpty())
			},
			Entry("missing hash sign", "PR\r\n\r\nFixes: 1 issue"),
			Entry("space after other repo path", "PR closed: org/repo #1 issue"),
			Entry("missing number of the issue", "PR\r\nresolve: org/my-repo#"),
			Entry("missing whole link", "this PR fixes issues"),
			Entry("wrong keyword with correct other repo link", "PR verifies org/repo#1 issue"),
			Entry("wrong keyword with correct issue link", "PR manages #1 issue"),
			Entry("additional words between keyword and issue link", "PR fixes bugs in #1"),
		)

		DescribeTable("should approve description that contains enough number of characters",
			func(desc string) {
				pr := &github.PullRequest{Body: utils.String(desc)}
				config := prsanitizer.PluginConfiguration{DescriptionContentLength: 15}
				msg := prsanitizer.CheckDescriptionLength(pr, config, log.NewTestLogger())
				Expect(msg).To(BeEmpty())
			},
			Entry("with fixes keyword but without link", "This PR fixes bugs"),
			Entry("description with minimal number of chars", "It fixes issues"),
		)

		DescribeTable("should NOT approve description that doesn't contain enough number of characters",
			func(desc string) {
				pr := &github.PullRequest{Body: utils.String(desc)}
				config := prsanitizer.PluginConfiguration{DescriptionContentLength: 15}
				msg := prsanitizer.CheckDescriptionLength(pr, config, log.NewTestLogger())
				Expect(msg).NotTo(BeEmpty())
			},
			Entry("with fixes keyword and issue link", "This PR fixes #1 issue"),
			Entry("description with one char less than the minimal number", "It fixes issue"),
		)
	})

})

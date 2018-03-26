package plugin_test

import (
	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper/plugin"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("File pattern features", func() {

	var it = func(filePattern, expectedRegex string) {
		parsed := plugin.ParseFilePatterns([]string{filePattern})
		Expect(parsed).To(ConsistOf(plugin.FilePattern{Regex: expectedRegex}))
	}

	Context("File pattern parsing", func() {

		table.DescribeTable(
			"should parse file patterns to regex",
			it,
			should().
				parsePattern("**/*Test.java").
				toRegex(`.*/[^/]*Test\.java$`),
			should().
				parsePattern("*/*Test.java").
				toRegex(`[^/]*/[^/]*Test\.java$`),
			should().
				parsePattern("*Test.java").
				toRegex(`.*Test\.java$`),
			should().
				parsePattern("pkg/**/*_test.go").
				toRegex(`pkg/.*/[^/]*_test\.go$`),
			should().
				parsePattern("vendor/").
				toRegex(`vendor/.*`),
			should().
				parsePattern("pkg/*/**/*_test.go").
				toRegex(`pkg/[^/]*/.*/[^/]*_test\.go$`),
			should().
				parsePattern("test_*.py").
				toRegex(`test_[^/]*\.py$`))

		It("it should extract regex", func() {
			// given
			regexDef := []string{"regex{{my-regex}}"}

			// when
			parsed := plugin.ParseFilePatterns(regexDef)

			// then
			Expect(parsed).To(ConsistOf(plugin.FilePattern{Regex: "my-regex"}))
		})
	})
})

type assertion func()
type filePatternProvider func() string

var patternAssertionMsg = "Should parse file pattern %s to regex %s"

func should() assertion {
	return assertion(func() {})
}

func (p assertion) parsePattern(filePattern string) filePatternProvider {
	return filePatternProvider(func() string {
		return filePattern
	})
}

func (f filePatternProvider) toRegex(expRegex string) table.TableEntry {
	return table.Entry(fmt.Sprintf(patternAssertionMsg, f(), expRegex), f(), expRegex)
}

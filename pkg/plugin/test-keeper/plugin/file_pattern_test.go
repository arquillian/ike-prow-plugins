package plugin_test

import (
	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper/plugin"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("File pattern features", func() {

	var it = func(filePattern, expectedRegexp string) {
		parsed := plugin.ParseFilePatterns([]string{filePattern})
		Expect(parsed).To(ConsistOf(plugin.FilePattern{Regexp: expectedRegexp}))
	}

	Context("File pattern parsing", func() {

		table.DescribeTable(
			"should parse file patterns to regexp",
			it,
			should().
				parsePattern("**/*Test.java").
				toRegexp(`.*/[^/]*Test\.java$`),
			should().
				parsePattern("*/*Test.java").
				toRegexp(`[^/]*/[^/]*Test\.java$`),
			should().
				parsePattern("*Test.java").
				toRegexp(`.*Test\.java$`),
			should().
				parsePattern("pkg/**/*_test.go").
				toRegexp(`pkg/.*/[^/]*_test\.go$`),
			should().
				parsePattern("vendor/").
				toRegexp(`vendor/.*`),
			should().
				parsePattern("pkg/*/**/*_test.go").
				toRegexp(`pkg/[^/]*/.*/[^/]*_test\.go$`),
			should().
				parsePattern("test_*.py").
				toRegexp(`test_[^/]*\.py$`))

		It("should extract regexp", func() {
			// given
			regexpDef := []string{"regex{{my-regexp}}"}

			// when
			parsed := plugin.ParseFilePatterns(regexpDef)

			// then
			Expect(parsed).To(ConsistOf(plugin.FilePattern{Regexp: "my-regexp"}))
		})
	})
})

type assertion func()
type filePatternProvider func() string

var patternAssertionMsg = "Should parse file pattern %s to regexp %s"

func should() assertion {
	return assertion(func() {})
}

func (p assertion) parsePattern(filePattern string) filePatternProvider {
	return filePatternProvider(func() string {
		return filePattern
	})
}

func (f filePatternProvider) toRegexp(expRegexp string) table.TableEntry {
	return table.Entry(fmt.Sprintf(patternAssertionMsg, f(), expRegexp), f(), expRegexp)
}

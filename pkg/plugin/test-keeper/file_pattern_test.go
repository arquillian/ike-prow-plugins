package test_keeper_test

import (
	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("File pattern features", func() {

	var assertThat = func(filePattern, expectedRegexp string) {
		parsed := test_keeper.ParseFilePatterns([]string{filePattern})
		Expect(parsed).To(ConsistOf(test_keeper.FilePattern{Regexp: expectedRegexp}))
	}

	Context("File pattern parsing", func() {

		table.DescribeTable(
			"should parse file patterns to regexp",
			assertThat,
			pattern("**/*Test.java").isParsedToRegexp(`.*/[^/]*Test\.java$`),
			pattern("*/*Test.java").isParsedToRegexp(`[^/]*/[^/]*Test\.java$`),
			pattern("*Test.java").isParsedToRegexp(`.*Test\.java$`),
			pattern("pkg/**/*_test.go").isParsedToRegexp(`pkg/.*/[^/]*_test\.go$`),
			pattern("vendor/").isParsedToRegexp(`vendor/.*`),
			pattern("pkg/*/**/*_test.go").isParsedToRegexp(`pkg/[^/]*/.*/[^/]*_test\.go$`),
			pattern("test_*.py").isParsedToRegexp(`test_[^/]*\.py$`))

		It("should extract regexp", func() {
			// given
			regexpDef := []string{"regex{{my-regexp}}"}

			// when
			parsed := test_keeper.ParseFilePatterns(regexpDef)

			// then
			Expect(parsed).To(ConsistOf(test_keeper.FilePattern{Regexp: "my-regexp"}))
		})
	})
})

type filePatternProvider func() string

var patternAssertionMsg = "Should parse file pattern %s to regexp %s"

func pattern(filePattern string) filePatternProvider {
	return filePatternProvider(func() string {
		return filePattern
	})
}

func (f filePatternProvider) isParsedToRegexp(expRegexp string) table.TableEntry {
	return table.Entry(fmt.Sprintf(patternAssertionMsg, f(), expRegexp), f(), expRegexp)
}

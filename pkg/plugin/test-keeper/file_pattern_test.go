package testkeeper_test

import (
	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("File pattern features", func() {

	Context("File pattern parsing", func() {

		It("should extract regexp", func() {
			// given
			regexpDef := []string{"regex{{my-regexp}}"}

			// when
			parsed := testkeeper.ParseFilePatterns(regexpDef)

			// then
			Expect(parsed).To(ConsistOf(testkeeper.FilePattern{Regexp: "my-regexp"}))
		})
	})

	Context("File pattern matching", func() {

		var assertThat = func(file, pattern string) {
			parsed := testkeeper.ParseFilePatterns([]string{pattern})
			Expect(parsed.Matches(file)).To(BeTrue())
		}

		table.DescribeTable(
			"should parse file patterns to regexp",
			assertThat,
			file("src/main/resources/Anyfile").matches("**/Anyfile"),
			file("Anyfile").matches("**/**/Anyfile"),
			file("src/Anyfile").matches("*/Anyfile"),
			file("src/test/resources/Anyfile").matches("src/**/Anyfile"),
			file("src/test/resources/Anyfile").matches("*/Anyfile"), // FIXME this should fail as it's single directory
			file("Anyfile").matches("**/Anyfile"),
			file("test/directory/Anyfile").matches("*/Anyfile"),
			file("test/multiple/directory/Anyfile").matches("test/multiple/*/Anyfile"),
			file("Anyfile").matches("Anyfile"),
			file("test_case.py").matches("**/test*.py"),
			file("pkg/test/test_case.py").matches("**/test*.py"),
		)

	})
})

type filePatternProvider func() string

var patternAssertionMsg = "Should match file %s to expression %s"

func file(fileName string) filePatternProvider {
	return filePatternProvider(func() string {
		return fileName
	})
}

func (f filePatternProvider) matches(simplifiedRegExp string) table.TableEntry {
	return table.Entry(fmt.Sprintf(patternAssertionMsg, f(), simplifiedRegExp), f(), simplifiedRegExp)
}

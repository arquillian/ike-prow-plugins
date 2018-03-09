package plugin_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/arquillian/ike-prow-plugins/plugin/test-keeper/plugin"
	"github.com/onsi/ginkgo/extensions/table"
	"fmt"
)

var _ = Describe("Test Matcher features", func() {

	Context("Test matcher loading", func() {

		It("should load test matchers for java, go and typescript", func() {
			// when
			matchers := LoadMatchers(TestKeeperConfiguration{}, func() []string {
				return []string{"HTML", "Java", "Go", "TypeScript"}
			})

			// then
			Expect(matchers).To(HaveLen(3))
			Expect(matchers).To(ConsistOf(JavaMatcher, GoMatcher, TypeScriptMatcher))
		})

		It("should load default matcher when no supported language is set", func() {
			// when
			matchers := LoadMatchers(TestKeeperConfiguration{}, func() []string {
				return []string{"html, bash, haskel"}
			})

			// then
			Expect(matchers).To(HaveLen(1))
			Expect(matchers).To(ContainElement(DefaultMatcher))
		})
	})

	Context("Predefined matchers regex check", func() {

		table.DescribeTable("DefaultMatcher should return true for files that contain word 'test'",
			func(matcher FileNameMatcher, file string, shouldMatch bool) {
				Expect(matcher.Matches(file)).To(Equal(shouldMatch))
			},
			// when non test file then false
			createEntry(DefaultMatcher, "/path/to/MyAssertion.java", false),

			// when test file then true
			createEntry(DefaultMatcher, "/path/to/my_test.go", true),
			createEntry(DefaultMatcher, "/path/to/my.test.js", true),
			createEntry(DefaultMatcher, "/path/test/any.test.tsx", true),
			createEntry(DefaultMatcher, "/path/to/my_test.py", true),
			createEntry(DefaultMatcher, "/path/test/MyTestCase.groovy", true),
			createEntry(DefaultMatcher, "/path/test/TestAnything.java", true),
		)


		table.DescribeTable("JavaMatcher should return true only when matches Java test file",
			func(matcher FileNameMatcher, file string, shouldMatch bool) {
				Expect(matcher.Matches(file)).To(Equal(shouldMatch))
			},
			// when non test file then false
			createEntry(JavaMatcher, "/path/to/MyAssertion.java", false),
			createEntry(JavaMatcher, "/path/test/MyAssertions.java", false),

			// when test file then true
			createEntry(JavaMatcher, "/path/test/TestAnything.java", true),
			createEntry(JavaMatcher, "/path/test/MyTest.java", true),
			createEntry(JavaMatcher, "/path/test/MyTests.java", true),
			createEntry(JavaMatcher, "/path/test/MyTestCase.java", true),
		)

		table.DescribeTable("GoMatcher should return true only when matches Go test file",
			func(matcher FileNameMatcher, file string, shouldMatch bool) {
				Expect(matcher.Matches(file)).To(Equal(shouldMatch))
			},
			// when non test file then false
			createEntry(GoMatcher, "/path/to/my_assertion.go", false),
			createEntry(GoMatcher, "/path/test/my_assertion.go", false),
			createEntry(GoMatcher, "/path/test/test_anything.go", false),
			createEntry(GoMatcher, "/path/test/anytest.go", false),

			// when test file then true
			createEntry(GoMatcher, "/path/to/my_test.go", true),
		)

		table.DescribeTable("JavaScriptMatcher should return true only when matches JavaScript test file",
			func(matcher FileNameMatcher, file string, shouldMatch bool) {
				Expect(matcher.Matches(file)).To(Equal(shouldMatch))
			},
			// when non test file then false
			createEntry(JavaScriptMatcher, "/path/to/my.assertion.js", false),
			createEntry(JavaScriptMatcher, "/path/test/my.assertion.js", false),
			createEntry(JavaScriptMatcher, "/path/test/test.anything.js", false),

			// when test file then true
			createEntry(JavaScriptMatcher, "/path/test/any.test.js", true),
			createEntry(JavaScriptMatcher, "/path/to/my.test.js", true),
			createEntry(JavaScriptMatcher, "/path/to/my.spec.js", true),
		)

		table.DescribeTable("TypeScriptMatcher should return true only when matches TypeScript test file",
			func(matcher FileNameMatcher, file string, shouldMatch bool) {
				Expect(matcher.Matches(file)).To(Equal(shouldMatch))
			},
			// when non test file then false
			createEntry(TypeScriptMatcher, "/path/to/my.assertion.ts", false),
			createEntry(TypeScriptMatcher, "/path/test/my.assertion.tsx", false),
			createEntry(TypeScriptMatcher, "/path/test/test.anything.ts", false),

			// when test file then true
			createEntry(TypeScriptMatcher, "/path/test/any.test.tsx", true),
			createEntry(TypeScriptMatcher, "/path/to/my.test.ts", true),
			createEntry(TypeScriptMatcher, "/path/to/my.spec.ts", true),
		)

		table.DescribeTable("PythonMatcher should return true only when matches Python test file",
			func(matcher FileNameMatcher, file string, shouldMatch bool) {
				Expect(matcher.Matches(file)).To(Equal(shouldMatch))
			},
			// when non test file then false
			createEntry(PythonMatcher, "/path/to/my_assertion.py", false),
			createEntry(PythonMatcher, "/path/test/my_assertion.py", false),

			// when test file then true
			createEntry(PythonMatcher, "/path/test/test_anything.py", true),
			createEntry(PythonMatcher, "/path/to/my_test.py", true),
		)

		table.DescribeTable("GroovyMatcher should return true only when matches Groovy test file",
			func(matcher FileNameMatcher, file string, shouldMatch bool) {
				Expect(matcher.Matches(file)).To(Equal(shouldMatch))
			},
			// when non test file then false
			createEntry(GroovyMatcher, "/path/to/MyAssertion.groovy", false),
			createEntry(GroovyMatcher, "/path/test/MyAssertions.groovy", false),

			// when test file then true
			createEntry(GroovyMatcher, "/path/test/TestAnything.groovy", true),
			createEntry(GroovyMatcher, "/path/test/MyTest.groovy", true),
			createEntry(GroovyMatcher, "/path/test/MyTests.groovy", true),
			createEntry(GroovyMatcher, "/path/test/MyTestCase.groovy", true),
		)
	})
})

func createEntry(matcher FileNameMatcher, file string, shouldMatch bool) table.TableEntry {
	msg := "Test matcher should%s match the file %s, but it did%s."
	if shouldMatch {
		msg = fmt.Sprintf(msg, "", file, " NOT")
	} else {
		msg = fmt.Sprintf(msg, " NOT", file, "")
	}

	return table.Entry(msg, matcher, file, shouldMatch)
}

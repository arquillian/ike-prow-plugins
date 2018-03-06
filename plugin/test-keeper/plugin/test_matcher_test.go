package plugin_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"fmt"
	. "github.com/arquillian/ike-prow-plugins/plugin/test-keeper/plugin"
)

var _ = Describe("Test Matcher features", func() {

	Context("Test matcher loading", func() {

		It("should load test matcher from config", func() {
			// when
			matcher := LoadMatcherFromConfig(Logger, []byte("tests_pattern: .*my*.|test.go|pattern.js"))

			// then
			Expect(matcher.TestRegex).To(Equal(".*my*.|test.go|pattern.js"))
		})

		It("should load test matchers for java, go and typescript", func() {
			// when
			matchers := LoadTestMatchers([]string{"HTML", "Java", "Go", "TypeScript"})

			// then
			Expect(len(matchers)).To(Equal(3))
			Expect(matchers).To(ContainElement(JavaMatcher))
			Expect(matchers).To(ContainElement(GoMatcher))
			Expect(matchers).To(ContainElement(TypeScriptMatcher))
		})

		It("should load default matcher when no supported language is set", func() {
			// when
			matchers := LoadTestMatchers([]string{"html, bash, haskel"})

			// then
			Expect(len(matchers)).To(Equal(1))
			Expect(matchers).To(ContainElement(DefaultMatcher))
		})
	})

	Context("Predefined matchers regex check", func() {

		It("DefaultMatcher should return true for files that contains word \"test\"", func() {
			// when non test file then false
			verifyThatMatcherMatchesCorrectFile(JavaMatcher, "/path/to/MyAssertion.java", false)
			verifyThatMatcherMatchesCorrectFile(JavaMatcher, "/path/test/MyAssertions.java", false)

			// when test file then true
			verifyThatMatcherMatchesCorrectFile(JavaMatcher, "/path/test/TestAnything.java", true)
			verifyThatMatcherMatchesCorrectFile(JavaMatcher, "/path/test/MyTest.java", true)
			verifyThatMatcherMatchesCorrectFile(JavaMatcher, "/path/test/MyTests.java", true)
			verifyThatMatcherMatchesCorrectFile(JavaMatcher, "/path/test/MyTestCase.java", true)
		})

		It("JavaMatcher should return true only when matches Java test file", func() {
			// when non test file then false
			verifyThatMatcherMatchesCorrectFile(JavaMatcher, "/path/to/MyAssertion.java", false)
			verifyThatMatcherMatchesCorrectFile(JavaMatcher, "/path/test/MyAssertions.java", false)

			// when test file then true
			verifyThatMatcherMatchesCorrectFile(JavaMatcher, "/path/test/TestAnything.java", true)
			verifyThatMatcherMatchesCorrectFile(JavaMatcher, "/path/test/MyTest.java", true)
			verifyThatMatcherMatchesCorrectFile(JavaMatcher, "/path/test/MyTests.java", true)
			verifyThatMatcherMatchesCorrectFile(JavaMatcher, "/path/test/MyTestCase.java", true)
		})

		It("GoMatcher should return true only when matches Go test file", func() {
			// when non test file then false
			verifyThatMatcherMatchesCorrectFile(GoMatcher, "/path/to/my_assertion.go", false)
			verifyThatMatcherMatchesCorrectFile(GoMatcher, "/path/test/my_assertion.go", false)
			verifyThatMatcherMatchesCorrectFile(GoMatcher, "/path/test/test_anything.go", false)
			verifyThatMatcherMatchesCorrectFile(GoMatcher, "/path/test/anytest.go", false)

			// when test file then true
			verifyThatMatcherMatchesCorrectFile(GoMatcher, "/path/to/my_test.go", true)
		})

		It("JavaScriptMatcher should return true only when matches JavaScript test file", func() {
			// when non test file then false
			verifyThatMatcherMatchesCorrectFile(JavaScriptMatcher, "/path/to/my.assertion.js", false)
			verifyThatMatcherMatchesCorrectFile(JavaScriptMatcher, "/path/test/my.assertion.js", false)
			verifyThatMatcherMatchesCorrectFile(JavaScriptMatcher, "/path/test/test.anything.js", false)

			// when test file then true
			verifyThatMatcherMatchesCorrectFile(JavaScriptMatcher, "/path/test/any.test.js", true)
			verifyThatMatcherMatchesCorrectFile(JavaScriptMatcher, "/path/to/my.test.js", true)
			verifyThatMatcherMatchesCorrectFile(JavaScriptMatcher, "/path/to/my.spec.js", true)
		})

		It("TypeScriptMatcher should return true only when matches TypeScript test file", func() {
			// when non test file then false
			verifyThatMatcherMatchesCorrectFile(TypeScriptMatcher, "/path/to/my.assertion.ts", false)
			verifyThatMatcherMatchesCorrectFile(TypeScriptMatcher, "/path/test/my.assertion.tsx", false)
			verifyThatMatcherMatchesCorrectFile(TypeScriptMatcher, "/path/test/test.anything.ts", false)

			// when test file then true
			verifyThatMatcherMatchesCorrectFile(TypeScriptMatcher, "/path/test/any.test.tsx", true)
			verifyThatMatcherMatchesCorrectFile(TypeScriptMatcher, "/path/to/my.test.ts", true)
			verifyThatMatcherMatchesCorrectFile(TypeScriptMatcher, "/path/to/my.spec.ts", true)
		})

		It("PythonMatcher should return true only when matches Python test file", func() {
			// when non test file then false
			verifyThatMatcherMatchesCorrectFile(PythonMatcher, "/path/to/my_assertion.py", false)
			verifyThatMatcherMatchesCorrectFile(PythonMatcher, "/path/test/my_assertion.py", false)

			// when test file then true
			verifyThatMatcherMatchesCorrectFile(PythonMatcher, "/path/test/test_anything.py", true)
			verifyThatMatcherMatchesCorrectFile(PythonMatcher, "/path/to/my_test.py", true)
		})

		It("GroovyMatcher should return true only when matches Groovy test file", func() {
			// when non test file then false
			verifyThatMatcherMatchesCorrectFile(GroovyMatcher, "/path/to/MyAssertion.groovy", false)
			verifyThatMatcherMatchesCorrectFile(GroovyMatcher, "/path/test/MyAssertions.groovy", false)

			// when test file then true
			verifyThatMatcherMatchesCorrectFile(GroovyMatcher, "/path/test/TestAnything.groovy", true)
			verifyThatMatcherMatchesCorrectFile(GroovyMatcher, "/path/test/MyTest.groovy", true)
			verifyThatMatcherMatchesCorrectFile(GroovyMatcher, "/path/test/MyTests.groovy", true)
			verifyThatMatcherMatchesCorrectFile(GroovyMatcher, "/path/test/MyTestCase.groovy", true)
		})
	})
})

func verifyThatMatcherMatchesCorrectFile(matcher TestMatcher, file string, shouldMatch bool) {
	msg := "Test matcher should%s match the file %s, but it did%s."
	if shouldMatch {
		msg = fmt.Sprintf(msg, "", file, " NOT")
	} else {
		msg = fmt.Sprintf(msg, " NOT", file, "")
	}

	Expect(matcher.IsTest(file)).To(Equal(shouldMatch), msg)
}

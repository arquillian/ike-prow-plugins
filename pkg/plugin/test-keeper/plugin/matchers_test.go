package plugin_test

import (
	"fmt"

	. "github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper/plugin"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test Matcher features", func() {

	Context("Test matcher loading", func() {

		It("should load default matchers when no pattern is defined", func() {
			// given
			emptyConfiguration := TestKeeperConfiguration{}

			// when
			matchers := LoadMatcher(emptyConfiguration)

			// then
			Expect(matchers).To(Equal(DefaultMatchers))
		})

		It("should load defined inclusion pattern without default language specific matchers", func() {
			// given
			configurationWithInclusionPattern := TestKeeperConfiguration{Inclusion: `*IT.java|*TestCase.java`}
			firstRegex := func(matcher TestMatcher) string {
				return matcher.Inclusion[0].Regex
			}

			// when
			matchers := LoadMatcher(configurationWithInclusionPattern)

			// then
			Expect(matchers.Inclusion).To(HaveLen(1))
			Expect(matchers).To(WithTransform(firstRegex, Equal("*IT.java|*TestCase.java")))
		})
	})

	assertFileMatchers := func(matchers []FileNamePattern, file string, shouldMatch bool) {
		Expect(Matches(matchers, file)).To(Equal(shouldMatch))
	}

	Context("Predefined exclusion regex check", func() {

		table.DescribeTable("DefaultMatchers should exclude common build tools",
			assertFileMatchers,

			createEntry(DefaultMatchers.Exclusion, "src/github.com/arquillian/ike-prow-plugins/Makefile", true),
			createEntry(DefaultMatchers.Exclusion, "src/main/java/pom.xml", true),
			createEntry(DefaultMatchers.Exclusion, "mvnw", true),
			createEntry(DefaultMatchers.Exclusion, "mvnw.cmd", true),
			createEntry(DefaultMatchers.Exclusion, "mvnw.bat", true),
			createEntry(DefaultMatchers.Exclusion, "build.gradle", true),
			createEntry(DefaultMatchers.Exclusion, "gulpfile.js", true),
		)

		table.DescribeTable("DefaultMatchers should exclude common config files",
			assertFileMatchers,

			createEntry(DefaultMatchers.Exclusion, ".nvmrc", true),
			createEntry(DefaultMatchers.Exclusion, ".htmlhintrc", true),
			createEntry(DefaultMatchers.Exclusion, ".stylelintrc", true),
			createEntry(DefaultMatchers.Exclusion, ".editorconfig", true),
			createEntry(DefaultMatchers.Exclusion, "protractor.config.js", true),
			createEntry(DefaultMatchers.Exclusion, "protractorEE.config.js", true),
			createEntry(DefaultMatchers.Exclusion, "project/js/config/karma.conf.js", true),
			createEntry(DefaultMatchers.Exclusion, "project/js/config/tsconfig.json", true),
			createEntry(DefaultMatchers.Exclusion, "requirements.txt", true),
			createEntry(DefaultMatchers.Exclusion, "gulpfile.js", true),
			createEntry(DefaultMatchers.Exclusion, "vendor/github.com/arquillian/ike-prow-pugins/plugin_test.go", true),
		)

		table.DescribeTable("DefaultMatchers should exclude common .ignore files",
			assertFileMatchers,

			createEntry(DefaultMatchers.Exclusion, ".gitignore", true),
			createEntry(DefaultMatchers.Exclusion, ".dockerignore", true),
			createEntry(DefaultMatchers.Exclusion, ".stylelintignore", true),
		)

		table.DescribeTable("DefaultMatchers should exclude common documentation files",
			assertFileMatchers,

			createEntry(DefaultMatchers.Exclusion, "README.adoc", true),
			createEntry(DefaultMatchers.Exclusion, "README.asciidoc", true),
			createEntry(DefaultMatchers.Exclusion, "testing.md", true),
			createEntry(DefaultMatchers.Exclusion, "testing.txt", true),
			createEntry(DefaultMatchers.Exclusion, "LICENSE", true),
			createEntry(DefaultMatchers.Exclusion, "CODEOWNERS", true),
		)

		table.DescribeTable("DefaultMatchers should exclude ui assets",
			assertFileMatchers,

			createEntry(DefaultMatchers.Exclusion, "style.sass", true),
			createEntry(DefaultMatchers.Exclusion, "style.css", true),
			createEntry(DefaultMatchers.Exclusion, "style.less", true),
			createEntry(DefaultMatchers.Exclusion, "style.scss", true),
			createEntry(DefaultMatchers.Exclusion, "meme.png", true),
			createEntry(DefaultMatchers.Exclusion, "chart.svg", true),
			createEntry(DefaultMatchers.Exclusion, "photo.jpg", true),
			createEntry(DefaultMatchers.Exclusion, "pic.jpeg", true),
			createEntry(DefaultMatchers.Exclusion, "reaction.gif", true),
		)
	})

	Context("Predefined inclusion regex check", func() {

		table.DescribeTable("DefaultMatchers should return true for files that contain word 'test'",
			assertFileMatchers,
			// when non test file then false
			createEntry(DefaultMatchers.Inclusion, "/path/to/Test.java/MyAssertion.java", false),

			// when test file then true
			createEntry(DefaultMatchers.Inclusion, "/path/to/my_test.go", true),
			createEntry(DefaultMatchers.Inclusion, "/path/to/my.test.js", true),
			createEntry(DefaultMatchers.Inclusion, "/path/test/any.test.tsx", true),
			createEntry(DefaultMatchers.Inclusion, "/path/to/my_test.py", true),
			createEntry(DefaultMatchers.Inclusion, "/path/test/MyTestCase.groovy", true),
			createEntry(DefaultMatchers.Inclusion, "/path/test/TestAnything.java", true),
		)

		table.DescribeTable("javaTests should return true only when matches Java test file", assertFileMatchers,
			// when non test file then false
			createEntry(DefaultMatchers.Inclusion, "/path/to/Test.java/MyAssertion.java", false),
			createEntry(DefaultMatchers.Inclusion, "/path/test/MyAssertions.java", false),

			// when test file then true
			createEntry(DefaultMatchers.Inclusion, "/path/test/TestAnything.java", true),
			createEntry(DefaultMatchers.Inclusion, "/path/test/MyTest.java", true),
			createEntry(DefaultMatchers.Inclusion, "/path/test/MyTests.java", true),
			createEntry(DefaultMatchers.Inclusion, "/path/test/MyTestCase.java", true),
		)

		table.DescribeTable("goTests should return true only when matches Go test file",
			assertFileMatchers,
			// when non test file then false
			createEntry(DefaultMatchers.Inclusion, "/path/to/Test.java/my_assertion.go", false),
			createEntry(DefaultMatchers.Inclusion, "/path/test/my_assertion.go", false),
			createEntry(DefaultMatchers.Inclusion, "/path/test/test_anything.go", false),
			createEntry(DefaultMatchers.Inclusion, "/path/test/anytest.go", false),

			// when test file then true
			createEntry(DefaultMatchers.Inclusion, "/path/to/my_test.go", true),
		)

		table.DescribeTable("javascriptTests should return true only when matches JavaScript test file",
			assertFileMatchers,
			// when non test file then false
			createEntry(DefaultMatchers.Inclusion, "/path/to/test.go/my.assertion.js", false),
			createEntry(DefaultMatchers.Inclusion, "/path/test/my.assertion.js", false),
			createEntry(DefaultMatchers.Inclusion, "/path/test/test.anything.js", false),

			// when test file then true
			createEntry(DefaultMatchers.Inclusion, "/path/test/any.test.js", true),
			createEntry(DefaultMatchers.Inclusion, "/path/to/my.test.js", true),
			createEntry(DefaultMatchers.Inclusion, "/path/to/my.spec.js", true),
		)

		table.DescribeTable("typeScriptTests should return true only when matches TypeScript test file",
			assertFileMatchers,
			// when non test file then false
			createEntry(DefaultMatchers.Inclusion, "/path/to/test.go/my.assertion.ts", false),
			createEntry(DefaultMatchers.Inclusion, "/path/test/my.assertion.tsx", false),
			createEntry(DefaultMatchers.Inclusion, "/path/test/test.anything.ts", false),

			// when test file then true
			createEntry(DefaultMatchers.Inclusion, "/path/test/any.test.tsx", true),
			createEntry(DefaultMatchers.Inclusion, "/path/to/my.test.ts", true),
			createEntry(DefaultMatchers.Inclusion, "/path/to/my.spec.ts", true),
		)

		table.DescribeTable("pythonTests should return true only when matches Python test file",
			assertFileMatchers,
			// when non test file then false
			createEntry(DefaultMatchers.Inclusion, "/path/to/test.ts/my_assertion.py", false),
			createEntry(DefaultMatchers.Inclusion, "/path/test/my_assertion.py", false),

			// when test file then true
			createEntry(DefaultMatchers.Inclusion, "/path/test/test_anything.py", true),
			createEntry(DefaultMatchers.Inclusion, "/path/to/my_test.py", true),
		)

		table.DescribeTable("groovyTests should return true only when matches Groovy test file",
			assertFileMatchers,

			// when non test file then false
			createEntry(DefaultMatchers.Inclusion, "/path/to/test.py/MyAssertion.groovy", false),
			createEntry(DefaultMatchers.Inclusion, "/path/test/MyAssertions.groovy", false),

			// when test file then true
			createEntry(DefaultMatchers.Inclusion, "/path/test/TestAnything.groovy", true),
			createEntry(DefaultMatchers.Inclusion, "/path/test/MyTest.groovy", true),
			createEntry(DefaultMatchers.Inclusion, "/path/test/MyTests.groovy", true),
			createEntry(DefaultMatchers.Inclusion, "/path/test/MyTestCase.groovy", true),
		)
	})
})

func createEntry(matchers []FileNamePattern, file string, shouldMatch bool) table.TableEntry {
	msg := "Test matcher should%s match the file %s, but it did%s."
	if shouldMatch {
		msg = fmt.Sprintf(msg, "", file, " NOT")
	} else {
		msg = fmt.Sprintf(msg, " NOT", file, "")
	}

	return table.Entry(msg, matchers, file, shouldMatch)
}

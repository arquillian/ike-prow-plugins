package plugin_test

import (
	"fmt"

	. "github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper/plugin"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var (
	buildAssets = []string{
		"src/github.com/arquillian/ike-prow-plugins/Makefile", "src/main/java/pom.xml",
		"mvnw", "mvnw.cmd", "mvnw.bat", "build.gradle", "gulpfile.js",
		"vendor/github.com/arquillian/runner.go",
	}

	configFiles = []string{
		".nvmrc", ".htmlhintrc", ".stylelintrc", ".editorconfig",
		"protractor.config.js", "protractorEE.config.js", "project/js/config/karma.conf.js",
		"project/js/config/tsconfig.json", "requirements.txt", "gulpfile.js",
	}

	ignoreFiles = []string{
		".gitignore", ".dockerignore", ".dockerignore",
	}

	textFiles = []string{
		"README.adoc", "README.asciidoc", "CONTRIBUTIONS.md", "testing.txt",
		"LICENSE", "CODEOWNERS",
	}

	visualAssets = []string{
		"style.sass", "style.css", "style.less", "style.scss",
		"meme.png", "chart.svg", "photo.jpg", "pic.jpeg", "reaction.gif",
	}

	testSourceCode = []string{
		"/path/to/my.test.js", "/path/to/my.spec.js",
		"/path/test/any.test.tsx", "/path/to/my.test.ts", "/path/to/my.spec.ts",
		"/path/test/test_anything.py", "/path/to/my_test.py",
		"/path/test/TestAnything.groovy", "/path/test/MyTest.groovy", "/path/test/MyTests.groovy", "/path/test/MyTestCase.groovy",
		"/path/test/TestAnything.java", "/path/test/MyTest.java", "/path/test/MyTests.java",
		"/path/to/my_test.go",
	}

	regularSourceCode = []string{
		"/path/to/Test.java/MyAssertion.java", "/path/to/Test.java/MyAssertion.java",
		"/path/to/test.py/MyAssertion.groovy", "/path/test/MyAssertions.groovy",
		"/path/test/anytest.go", "/path/test/my_assertion.go", "/path/test/test_anything.go",
		"/path/to/test.go/my.assertion.js", "/path/test/my.assertion.js", "/path/test/test.anything.js",
		"/path/to/test.go/my.assertion.ts", "/path/test/my.assertion.tsx", "/path/test/test.anything.ts",
		"/path/test/my_assertion.py",
	}


	allNoTestFiles = func() []string {
		var all []string
		all = append(all, regularSourceCode...)
		all = append(all, buildAssets...)
		all = append(all, configFiles...)
		all = append(all, ignoreFiles...)
		all = append(all, textFiles...)
		all = append(all, visualAssets...)
		return all
	}()
)

var assertFileMatchers = func(matchers []FileNamePattern, file string, shouldMatch bool) {
	Expect(Matches(matchers, file)).To(Equal(shouldMatch))
}

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

	Context("Predefined exclusion regex check (DefaultMatchers)", func() {

		table.DescribeTable("should exclude common build tools",
			assertFileMatchers,
			matchingEntries(DefaultMatchers.Exclusion, buildAssets)...
		)


		table.DescribeTable("should exclude common config files",
			assertFileMatchers,
			matchingEntries(DefaultMatchers.Exclusion, configFiles)...

		)

		table.DescribeTable("should exclude common .ignore files",
			assertFileMatchers,
			matchingEntries(DefaultMatchers.Exclusion, ignoreFiles)...
		)

		table.DescribeTable("should exclude common documentation files",
			assertFileMatchers,
			matchingEntries(DefaultMatchers.Exclusion, textFiles)...
		)

		table.DescribeTable("should exclude ui assets",
			assertFileMatchers,
			matchingEntries(DefaultMatchers.Exclusion, visualAssets)...
		)
	})

	Context("Predefined inclusion regex check (DefaultMatchers)", func() {

		table.DescribeTable("should include common test naming conventions",
			assertFileMatchers,
			matchingEntries(DefaultMatchers.Inclusion, testSourceCode)...
		)

		table.DescribeTable("should not include other source files",
			assertFileMatchers,
			notMatchingEntries(DefaultMatchers.Inclusion, allNoTestFiles)...
		)
	})
})

func matchingEntries(patterns []FileNamePattern, files []string) []table.TableEntry {
	return entries(patterns, files, true)
}

func notMatchingEntries(patterns []FileNamePattern, files []string) []table.TableEntry {
	return entries(patterns, files, false)
}

func entries(patterns []FileNamePattern, files []string, shouldMatch bool) []table.TableEntry {
	entries := make([]table.TableEntry, len(files))

	for i, file := range files {
		entries[i] = createEntry(patterns, file, shouldMatch)
	}

	return entries
}

const msg = "Test matcher should%s match the file %s, but it did%s."

func createEntry(matchers []FileNamePattern, file string, shouldMatch bool) table.TableEntry {
	if shouldMatch {
		return table.Entry(fmt.Sprintf(msg, "", file, " NOT"), matchers, file, shouldMatch)
	}
	
	return table.Entry(fmt.Sprintf(msg, " NOT", file, ""), matchers, file, shouldMatch)
}
package testkeeper_test

import (
	"fmt"

	. "github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	buildAssets = []string{
		"src/github.com/arquillian/ike-prow-plugins/Makefile", ".make/Makefile.docker", ".make/test.mk",
		"src/main/java/pom.xml", "mvnw", "mvnw.cmd", "mvnw.bat", "build.gradle", "gulpfile.js", "Gruntfile.js",
		"gradlew", "gradlew.bat", "Rakefile",
		"glide.yaml", "glide.lock", "pom.xml", "package.json", "package-lock.json",
		"vendor/github.com/arquillian/runner.go", "Gopkg.toml", "Gopkg.lock",
		"docker-compose.yml", "docker-compose.dev.yml", "docker-compose.macos.yaml",
		"Dockerfile", "Dockerfile.builder", "Dockerfile.deploy",
	}

	configFiles = []string{
		".nvmrc", ".htmlhintrc", ".stylelintrc", ".editorconfig", "typedoc.json",
		"protractor.config.js", "protractorEE.config.js", "project/js/config/karma.conf.js",
		"project/js/config/tsconfig.json", "requirements.txt", "gulpfile.js", "tslint.json",
		"0034A06D9D9B0064CE8ADF6BF1747F4AD2306D93.gpg", "webpack.config.js", "pylint.rc",
		"codeship-services.yml", ".golint_exclude", ".gofmt_exclude", "pcp.repo", ".sass-lint.yml",
	}

	shellScripts = []string{
		"openshift-prod-cluster.sh", "test.bat", "cico_build_deploy.sh",
	}

	ignoreFiles = []string{
		".gitignore", ".dockerignore", ".stylelintignore",
	}

	textFiles = []string{
		"README.adoc", "README.asciidoc", "CONTRIBUTIONS.md", "testing.txt",
		"LICENSE", "CODEOWNERS",
	}

	visualAssets = []string{
		"style.sass", "style.css", "style.less", "style.scss",
		"meme.png", "chart.svg", "photo.jpg", "pic.jpeg", "reaction.gif", "image.bmp", "image.tiff",
		"index.html", "fav.ico", "index.shtml", "index.dhtml", "index.xhtml",
		"template.ejs", "vector.eps", "image.raw",
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
		all = append(all, shellScripts...)
		return all
	}()
)

var expectThatFile = func(matchers []FilePattern, file string, shouldMatch bool) {
	Expect(Matches(matchers, file)).To(Equal(shouldMatch))
}

var _ = Describe("Test Matcher features", func() {

	var DefaultMatchers, _ = LoadDefaultMatcher()

	Context("Test matcher loading", func() {

		It("should load default matchers when no pattern is defined", func() {
			// given
			emptyConfiguration := &PluginConfiguration{}

			// when
			matchers, err := LoadMatcher(emptyConfiguration)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(matchers).To(Equal(DefaultMatchers))
		})

		It("should load defined inclusion pattern without default language specific matchers", func() {
			// given
			configurationWithInclusionPattern := &PluginConfiguration{
				Inclusions: []string{`regex{{*IT.java|*TestCase.java}}`},
			}
			firstRegexp := func(matcher TestMatcher) string {
				return matcher.Inclusion[0].Regexp
			}

			// when
			matchers, err := LoadMatcher(configurationWithInclusionPattern)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(matchers.Inclusion).To(HaveLen(1))
			Expect(matchers).To(WithTransform(firstRegexp, Equal("*IT.java|*TestCase.java")))
		})
	})

	Context("Predefined exclusion regexp check (DefaultMatchers)", func() {

		DescribeTable("should exclude common build tools",
			expectThatFile,
			from(buildAssets).matches(DefaultMatchers.Exclusion),
		)

		DescribeTable("should exclude common config files",
			expectThatFile,
			from(configFiles).matches(DefaultMatchers.Exclusion),
		)

		DescribeTable("should exclude common .ignore files",
			expectThatFile,
			from(ignoreFiles).matches(DefaultMatchers.Exclusion),
		)

		DescribeTable("should exclude common documentation files",
			expectThatFile,
			from(textFiles).matches(DefaultMatchers.Exclusion),
		)

		DescribeTable("should exclude ui assets",
			expectThatFile,
			from(visualAssets).matches(DefaultMatchers.Exclusion),
		)

		DescribeTable("should exclude shell scripts",
			expectThatFile,
			from(shellScripts).matches(DefaultMatchers.Exclusion),
		)
	})

	Context("Predefined inclusion regexp check (DefaultMatchers)", func() {

		DescribeTable("should include common test naming conventions",
			expectThatFile,
			from(testSourceCode).matches(DefaultMatchers.Inclusion),
		)

		DescribeTable("should not include other source files",
			expectThatFile,
			from(allNoTestFiles).doesNotMatch(DefaultMatchers.Inclusion),
		)
	})
})

func from(files []string) filesProvider {
	return filesProvider(func() []string {
		return files
	})
}

type filesProvider func() []string

func (f filesProvider) matches(patterns []FilePattern) []TableEntry {
	return entries(patterns, f(), true)
}

func (f filesProvider) doesNotMatch(patterns []FilePattern) []TableEntry {
	return entries(patterns, f(), false)
}

func entries(patterns []FilePattern, files []string, shouldMatch bool) []TableEntry {
	entries := make([]TableEntry, len(files))

	for i, file := range files {
		entries[i] = createEntry(patterns, file, shouldMatch)
	}

	return entries
}

const msg = "Test matcher should%s match the file %s, but it did%s."

func createEntry(matchers []FilePattern, file string, shouldMatch bool) TableEntry {
	if shouldMatch {
		return Entry(fmt.Sprintf(msg, "", file, " NOT"), matchers, file, shouldMatch)
	}

	return Entry(fmt.Sprintf(msg, " NOT", file, ""), matchers, file, shouldMatch)
}

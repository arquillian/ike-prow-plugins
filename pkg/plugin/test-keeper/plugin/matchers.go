package plugin

import (
	"regexp"
)

// FileNamePattern contains regex that matches a test file
type FileNamePattern struct {
	Regex string
}

// Matches checks if the given string (representing path to a file) contains a substring that matches Regex stored in this matcher
func (matcher *FileNamePattern) Matches(filename string) bool {
	return regexp.MustCompile(matcher.Regex).MatchString(filename)
}

// TestMatcher holds definitions of patterns considered as test filenames (inclusions) and those which shouldn't be
// verified (exclusions)
type TestMatcher struct {
	Inclusion []FileNamePattern
	Exclusion []FileNamePattern
}

// MatchesInclusion checks if file name matches defined inclusion patterns
func (matcher *TestMatcher) MatchesInclusion(filename string) bool {
	return Matches(matcher.Inclusion, filename)
}

// MatchesExclusion checks if file name matches defined exclusion patterns
func (matcher *TestMatcher) MatchesExclusion(filename string) bool {
	return Matches(matcher.Exclusion, filename)
}

// Matches iterates over a slice of FileNamePattern and verifies if passed name matches any of the defined patterns
func Matches(matchers []FileNamePattern, filename string) bool {

	for _, matcher := range matchers {
		if matcher.Matches(filename) {
			return true
		}
	}

	return false
}

// DefaultMatchers is used when no other matcher is loaded
// It matches any string that contains either "test" or "Test"
var DefaultMatchers = TestMatcher{
	Inclusion: []FileNamePattern{javaTests, goTests, javascriptTests, typescriptTests, pythonTests, groovyTests},
	Exclusion: []FileNamePattern{
		buildToolsFileNameMatcher, buildToolsDirectoryNameMatcher,
		ciToolsFileNameMatcher, settingsFileNameMatcher,
		textAssetsFileNameMatcher, uiAssetsFileNameMatcher,
	},
}

var javaTests = FileNamePattern{
	Regex: `(Test[^/]*|IT|TestCase)\.java$`,
}

var goTests = FileNamePattern{
	Regex: `_test\.go$`,
}

var javascriptTests = FileNamePattern{
	Regex: `(test|spec)\.js$`,
}

var typescriptTests = FileNamePattern{
	Regex: `(test|spec)\.ts(x)?$`,
}

var pythonTests = FileNamePattern{
	Regex: `test[^/]*\.py$`,
}

var groovyTests = FileNamePattern{
	Regex: `(Test[^/]*|IT|TestCase)\.groovy$`,
}

var buildToolsFileNameMatcher = FileNamePattern{
	Regex: `(glide\.yaml|glide\.lock|pom\.xml|mvnw(\.cmd|\.bat)?|package\.json|glide\.yaml|build\.gradle|gradlew[\.bat]?|Makefile|gulpfile\.js|(G|g)emfile|requirements\.(in|txt))$`,
}

var buildToolsDirectoryNameMatcher = FileNamePattern{
	Regex: `gradle/|vendor/|.mvn/|node_modules/`,
}

var ciToolsFileNameMatcher = FileNamePattern{
	Regex: `\.travis\.yml|Jenkinsfile|\.gitlab-ci\.yml,|wercker\.yml|circle\.yml$`,
}

var uiAssetsFileNameMatcher = FileNamePattern{
	Regex: `(\.jpg|\.jpeg|\.png|\.ico|\.svg|\.gif|\.css|\.scss|\.sass|\.less)$`,
}

var textAssetsFileNameMatcher = FileNamePattern{
	Regex: `(\.md|\.txt|\.asciidoc|\.adoc|LICENSE|CODEOWNERS)$`,
}

var settingsFileNameMatcher = FileNamePattern{
	Regex: `(karma\.conf\.js|protractor.*\.conf(ig)?\.js|\..+ignore|\.editorconfig|\..+rc|tsconfig.*\.json|karma\.conf\.js|\.codecov\.yml|pylint\.rc)$`,
}

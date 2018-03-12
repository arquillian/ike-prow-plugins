package plugin

import (
	"regexp"
)

// TestKeeperConfiguration defines inclusion and exclusion patterns set of files will be matched against
// It's unmarshaled from test-keeper.yml configuration file
type TestKeeperConfiguration struct {
	Inclusion string `yaml:"test_pattern,omitempty"`
	Exclusion string `yaml:"exclusion,omitempty"`
}

// TestMatcher holds definitions of patterns considered as test filenames (inclusions) and those which shouldn't be
// verified (exclusions)
type TestMatcher struct {
	Inclusion []FileNamePattern
	Exclusion []FileNamePattern
}

// FileNamePattern contains regex that matches a test file
type FileNamePattern struct {
	Regex string
}

// DefaultMatchers is used when no other matcher is loaded
// It matches any string that contains either "test" or "Test"
var DefaultMatchers = TestMatcher{
	Inclusion: []FileNamePattern{javaTests, goTests, javascriptTests, typescriptTests, pythonTests, groovyTests},
	Exclusion: []FileNamePattern{buildToolsFileNameMatcher, ciToolsFileNameMatcher, textAssetsFileNameMatcher},
}

// Matches checks if the given string (representing path to a file) contains a substring that matches Regex stored in this matcher
func (matcher *FileNamePattern) Matches(filename string) bool {
	return regexp.MustCompile(matcher.Regex).MatchString(filename)
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

// LoadMatcher loads list of FileNamePattern either from the provided configuration or from languages retrieved from the given function
func LoadMatcher(config TestKeeperConfiguration) TestMatcher {
	var matcher TestMatcher

	if config.Inclusion != "" {
		matcher = TestMatcher{Inclusion: []FileNamePattern{{Regex: config.Inclusion}}}
	} else {
		matcher = DefaultMatchers
	}

	return matcher
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
	Regex: `pom\.xml|mvnw[\.cmd]?|\.mvn|package\.json|glide\.yaml|build\.gradle|gradlew[\.bat]?|gradle/|Makefile`,
}

var ciToolsFileNameMatcher = FileNamePattern{
	Regex: `\.travis\.yml|Jenkinsfile|\.gitlab-ci\.yml`,
}

var textAssetsFileNameMatcher = FileNamePattern{
	Regex: `(\.md|\.adoc|LICENSE)$`,
}

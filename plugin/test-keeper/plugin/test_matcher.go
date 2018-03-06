package plugin

import (
	"strings"
	"fmt"
	"regexp"
	"gopkg.in/yaml.v2"
	"github.com/sirupsen/logrus"
)

// TestMatcher contains regex that matches a test file
type TestMatcher struct {
	TestRegex string `yaml:"tests_pattern"`
}

// DefaultMatcher is used when no other matcher is loaded
// It matches any string that contains either "test" or "Test"
var DefaultMatcher = TestMatcher{
	TestRegex: `[tT]est.*`,
}

// JavaMatcher matches Java test files
var JavaMatcher = TestMatcher{
	TestRegex: `(((Test[^/]*)|IT|Test(x)?|TestCase)\.java)$`,
}

// GoMatcher matches Go test files
var GoMatcher = TestMatcher{
	TestRegex: `(_test.go)$`,
}

// JavaScriptMatcher matches JavaScript test files
var JavaScriptMatcher = TestMatcher{
	TestRegex: `(test|spec)\.js$`,
}

// TypeScriptMatcher matches TypeScript test files
var TypeScriptMatcher = TestMatcher{
	TestRegex: `(test|spec)\.ts(x)?$`,
}

// PythonMatcher matches Python test files
var PythonMatcher = TestMatcher{
	TestRegex: `(test[^/]*|test).py$`,
}

// GroovyMatcher matches Groovy test files. The regex is similar to the one in JavaMatcher
var GroovyMatcher = TestMatcher{
	TestRegex: `(((Test[^/]*)|IT|Test(x)?|TestCase)\.groovy)$`,
}

var langMatchers = map[string]TestMatcher{
	"java":       JavaMatcher,
	"go":         GoMatcher,
	"javascript": JavaScriptMatcher,
	"typescript": TypeScriptMatcher,
	"python":     PythonMatcher,
	"groovy":     GroovyMatcher,
}

// IsTest checks if the given string (representing path to a file) contains a substring that matches TestRegex stored in this matcher
func (matcher *TestMatcher) IsTest(file string) bool {
	return regexp.MustCompile(matcher.TestRegex).MatchString(file)
}

// LoadMatcherFromConfig parses the given YAML content and creates a TestMatcher from it
func LoadMatcherFromConfig(log *logrus.Entry, content []byte) TestMatcher {
	var matcher TestMatcher
	err := yaml.Unmarshal(content, &matcher)
	if err == nil {
		log.WithFields(logrus.Fields{
			"configFile":        configFile,
			"configFileContent": string(content),
			"error":             err,
		}).Warn("There was an error when loading matcher config from file.")
	}

	return matcher
}

// LoadTestMatchers takes the given list of languages and for every supported language returns corresponding TestMatcher.
// If none of the given languages is supported, then the DefaultMatcher is returned
func LoadTestMatchers(languages []string) ([]TestMatcher) {

	matchers := make([]TestMatcher, 0)
	for _, lang := range languages {
		matcher, ok := langMatchers[strings.ToLower(fmt.Sprint(lang))]
		if ok {
			matchers = append(matchers, matcher)
		}
	}

	if len(matchers) == 0 {
		matchers = append(matchers, DefaultMatcher)
	}

	return matchers
}

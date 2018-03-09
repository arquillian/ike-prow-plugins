package plugin

import (
	"regexp"
	"strings"
	"fmt"
)

type TestKeeperConfiguration struct {
	Inclusion string `yaml:"test_pattern,omitempty"`
}

// FileNameMatcher contains regex that matches a test file
type FileNameMatcher struct {
	Regex string
}

// DefaultMatcher is used when no other matcher is loaded
// It matches any string that contains either "test" or "Test"
var DefaultMatcher = FileNameMatcher{
	Regex: `[tT]est.*`,
}

// JavaMatcher matches Java test files
var JavaMatcher = FileNameMatcher{
	Regex: `(Test[^/]*|IT|TestCase)\.java`,
}

// GoMatcher matches Go test files
var GoMatcher = FileNameMatcher{
	Regex: `_test\.go$`,
}

// JavaScriptMatcher matches JavaScript test files
var JavaScriptMatcher = FileNameMatcher{
	Regex: `(test|spec)\.js$`,
}

// TypeScriptMatcher matches TypeScript test files
var TypeScriptMatcher = FileNameMatcher{
	Regex: `(test|spec)\.ts(x)?$`,
}

// PythonMatcher matches Python test files
var PythonMatcher = FileNameMatcher{
	Regex: `test[^/]*\.py$`,
}

// GroovyMatcher matches Groovy test files. The regex is similar to the one in JavaMatcher
var GroovyMatcher = FileNameMatcher{
	Regex: `(Test[^/]*|IT|TestCase)\.groovy$`,
}

var langMatchers = map[string]FileNameMatcher{
	"java":       JavaMatcher,
	"go":         GoMatcher,
	"javascript": JavaScriptMatcher,
	"typescript": TypeScriptMatcher,
	"python":     PythonMatcher,
	"groovy":     GroovyMatcher,
}

// Matches checks if the given string (representing path to a file) contains a substring that matches Regex stored in this matcher
func (matcher *FileNameMatcher) Matches(file string) bool {
	return regexp.MustCompile(matcher.Regex).MatchString(file)
}

func LoadMatchers(config TestKeeperConfiguration, getLanguages func() []string) []FileNameMatcher {
	var matchers []FileNameMatcher

	if config.Inclusion != "" {
		matchers = append(matchers, FileNameMatcher{Regex: config.Inclusion})
	} else {
		matchers = loadMatchers(getLanguages())
	}
	return matchers
}


// LoadTestMatchers takes the given list of languages and for every supported language returns corresponding FileNameMatcher.
// If none of the given languages is supported, then the DefaultMatcher is returned
func loadMatchers(languages []string) ([]FileNameMatcher) {

	matchers := make([]FileNameMatcher, 0)
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

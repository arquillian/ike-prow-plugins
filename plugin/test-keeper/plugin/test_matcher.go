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

var defaultMatcher = TestMatcher{
	TestRegex: `[tT]est.*`,
}

var javaMatcher = TestMatcher{
	TestRegex: `((IT|Test|TestCase)\.java)$`,
}

var goMatcher = TestMatcher{
	TestRegex: `(_test.go)$`,
}

var javaScriptMatcher = TestMatcher{
	TestRegex: `(test|spec)\.js$`,
}

var typeScriptMatcher = TestMatcher{
	TestRegex: `(test|spec)\.ts(x)?$`,
}

var pythonMatcher = TestMatcher{
	TestRegex: `(test[^/]*|test).py$`,
}

var groovyMatcher = TestMatcher{
	TestRegex: `((IT|Test|TestCase)\.groovy)$`,
}

var langMatchers = map[string]TestMatcher{
	"java":       javaMatcher,
	"go":         goMatcher,
	"javascript": javaScriptMatcher,
	"typescript": typeScriptMatcher,
	"python":     pythonMatcher,
	"groovy":     groovyMatcher,
}

func (matcher *TestMatcher) isTest(file string) bool {
	return regexp.MustCompile(matcher.TestRegex).MatchString(file)
}

func loadMatcherFromConfig(log *logrus.Entry, content []byte) TestMatcher {
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

func loadTestMatchers(languages []string) ([]TestMatcher) {

	matchers := make([]TestMatcher, 0)
	for _, lang := range languages {
		matcher, ok := langMatchers[strings.ToLower(fmt.Sprint(lang))]
		if ok {
			matchers = append(matchers, matcher)
		}
	}

	if len(matchers) == 0 {
		matchers = append(matchers, defaultMatcher)
	}

	return matchers
}

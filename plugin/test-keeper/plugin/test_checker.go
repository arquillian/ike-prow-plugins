package plugin

import (
	"github.com/sirupsen/logrus"
	"github.com/arquillian/ike-prow-plugins/scm"
)

// TestChecker is using plugin.TestMatcher to figure out if the given commit affects any test file
// The plugin.TestMatcher is loaded either from test-keeper.yaml file or from set of default matchers based on the languages using in the related project
type TestChecker struct {
	Log           *logrus.Entry
	CommitService scm.CommitScmService
}

const configFile = "test-keeper.yaml"

// IsAnyTestPresent checks if a commit affects any test file
func (checker *TestChecker) IsAnyTestPresent() (bool, error) {

	matchers := checker.loadMatchers()

	checker.Log.Infof("Checking for tests")
	files, e := checker.CommitService.GetAffectedFiles()
	if e != nil {
		return false, e
	}

	for _, matcher := range matchers {
		for _, file := range files {
			checker.Log.Infof("%q: %q", file.Name, file.Status)

			if matcher.IsTest(file.Name) {
				return true, nil
			}
		}
	}

	return false, nil
}

func (checker *TestChecker) loadMatchers() []TestMatcher {
	var matchers []TestMatcher

	config := checker.CommitService.GetRawFile(configFile)
	if config != nil {
		matcher := LoadMatcherFromConfig(checker.Log, config)
		if matcher.TestRegex != "" {
			matchers = append(matchers, matcher)
		}
	}

	if len(matchers) == 0 {
		languages, e := checker.CommitService.GetRepoService().GetRepoLanguages()
		if e != nil {
			matchers = append(matchers, DefaultMatcher)
		} else {
			matchers = LoadTestMatchers(languages)
		}
	}
	return matchers
}

package plugin

import (
	"github.com/sirupsen/logrus"
	"github.com/arquillian/ike-prow-plugins/scm"
)


// TestChecker is using plugin.TestMatcher to figure out if the given commit affects any test file
// The plugin.TestMatcher is loaded either from test-keeper.yaml file or from set of default matchers based on the languages using in the related project
type TestChecker struct {
	log           *logrus.Entry
	commitService *scm.CommitScmService
}

const configFile = "test-keeper.yaml"

// isAnyTestPresent checks if a commit affects any test file
func (checker *TestChecker) isAnyTestPresent() (bool, error) {

	matchers := checker.loadMatchers()

	checker.log.Infof("Checking for tests")
	files, e := checker.commitService.GetAffectedFiles()
	if e != nil {
		return false, e
	}

	for _, matcher := range matchers {
		for _, file := range files {
			checker.log.Infof("%q: %q", file.Name, file.Status)

			if matcher.isTest(file.Name) {
				return true, nil
			}
		}
	}

	return false, nil
}

func (checker *TestChecker) loadMatchers() []TestMatcher {
	var matchers []TestMatcher

	config := checker.commitService.GetRawFile(configFile)
	if config != nil {
		matcher := loadMatcherFromConfig(checker.log, config)
		if matcher.TestRegex != "" {
			matchers = append(matchers, matcher)
		}
	}

	if len(matchers) == 0 {
		languages, e := checker.commitService.RepoService.GetRepoLanguages()
		if e != nil {
			matchers = append(matchers, defaultMatcher)
		} else {
			matchers = loadTestMatchers(languages)
		}
		logrus.Info("Using test matchers: ", matchers)
	} else {
		logrus.Info("Using test matcher from file: ", matchers)
	}
	return matchers
}

package plugin

import (
	"github.com/sirupsen/logrus"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"errors"
)

// TestChecker is using plugin.FileNamePattern to figure out if the given commit affects any test file
// The plugin.FileNamePattern is loaded either from test-keeper.yaml file or from set of default matchers based on the languages using in the related project
type TestChecker struct {
	Log               *logrus.Entry
	TestKeeperMatcher TestMatcher
}

// IsAnyNotExcludedFileTest checks if a commit affects any test file
func (checker *TestChecker) IsAnyNotExcludedFileTest(files []scm.ChangedFile) (bool, error) {
	checker.Log.Infof("Checking for tests")

	remainingNoTestFiles := false
	for _, file := range files {
		if file.Name == "" {
			return false, errors.New("can't have empty file name")
		}
		excluded := checker.TestKeeperMatcher.MatchesExclusion(file.Name)
		if !excluded {
			if checker.TestKeeperMatcher.MatchesInclusion(file.Name) {
				return true, nil
			}
			remainingNoTestFiles = true
		}
	}
	return !remainingNoTestFiles, nil
}

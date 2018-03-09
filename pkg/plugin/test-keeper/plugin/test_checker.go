package plugin

import (
	"github.com/sirupsen/logrus"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
)

// TestChecker is using plugin.FileNameMatcher to figure out if the given commit affects any test file
// The plugin.FileNameMatcher is loaded either from test-keeper.yaml file or from set of default matchers based on the languages using in the related project
type TestChecker struct {
	Log          *logrus.Entry
	TestMatchers []FileNameMatcher
}

// IsAnyTestPresent checks if a commit affects any test file
func (checker *TestChecker) IsAnyTestPresent(files []scm.ChangedFile) (bool, error) {

	checker.Log.Infof("Checking for tests")

	for _, matcher := range checker.TestMatchers {
		for _, file := range files {
			checker.Log.Infof("%q: %q", file.Name, file.Status)

			if matcher.Matches(file.Name) {
				return true, nil
			}
		}
	}

	return false, nil
}

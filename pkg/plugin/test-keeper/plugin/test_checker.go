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

// TestKeeperConfiguration defines inclusion and exclusion patterns set of files will be matched against
// It's unmarshaled from test-keeper.yml configuration file
type TestKeeperConfiguration struct {
	Inclusion string `yaml:"test_pattern,omitempty"`
	Exclusion string `yaml:"exclusion,omitempty"`
	// TODO combine_defaults: *true|false
}

// LoadMatcher loads list of FileNamePattern either from the provided configuration or from languages retrieved from the given function
func LoadMatcher(config TestKeeperConfiguration) TestMatcher {
	var matcher = DefaultMatchers

	if config.Inclusion != "" {
		matcher.Inclusion = []FileNamePattern{{Regex: config.Inclusion}}
	}

	if config.Exclusion != "" {
		matcher.Exclusion = []FileNamePattern{{Regex: config.Exclusion}}
	}

	return matcher
}

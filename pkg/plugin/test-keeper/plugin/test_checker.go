package plugin

import (
	"errors"

	"github.com/arquillian/ike-prow-plugins/pkg/scm"
)

// TestChecker is using plugin.FileNamePattern to figure out if the given commit affects any test file
// The plugin.FileNamePattern is loaded either from test-keeper.yaml file or from set of default matchers based on the languages using in the related project
type TestChecker struct {
	TestKeeperMatcher TestMatcher
}

// FileCategories holds information about the total files coming in the changeset, legit files (those which are excluded from test verification)
// and Tests
type FileCategories struct {
	Total, Legit, Tests int
	Files               *[]scm.ChangedFile
}

// OnlyLegitFiles indicates if changeset contains only files which are excluded from test verification
func (f *FileCategories) OnlyLegitFiles() bool {
	return f.Total > 0 && f.Legit == f.Total
}

// TestsExist answers if any test files are found
func (f *FileCategories) TestsExist() bool {
	return f.Tests > 0
}

// NewFileCategories creates new instance of FileCategories struct with files populated
func NewFileCategories(files []scm.ChangedFile) FileCategories {
	return FileCategories{Files: &files, Total: len(files)}
}

// CategorizeFiles counts files in the changeset which are tests (included files) and should not be considered for
// verification (excluded). When first test is found it stops, as this is enough to unblock PR
func (checker *TestChecker) CategorizeFiles(files []scm.ChangedFile) (FileCategories, error) {
	categories := NewFileCategories(files)
	for _, file := range files {
		if file.Name == "" {
			return categories, errors.New("can't have empty file name")
		}
		excluded := checker.TestKeeperMatcher.MatchesExclusion(file.Name)
		if !excluded {
			if checker.TestKeeperMatcher.MatchesInclusion(file.Name) {
				categories.Tests++
				return categories, nil // As we found the first test and we don't care about the amount of them, we can return
			}

		} else {
			categories.Legit++
		}
	}
	return categories, nil
}

// LoadMatcher loads list of FileNamePattern either from the provided configuration or from languages retrieved from the given function
func LoadMatcher(config TestKeeperConfiguration) TestMatcher {
	var matcher = DefaultMatchers

	if config.Inclusion != "" {
		matcher.Inclusion = []FileNamePattern{{Regex: config.Inclusion}}
		if config.Combine {
			matcher.Inclusion = append(matcher.Inclusion, DefaultMatchers.Inclusion...)
		}
	}

	if config.Exclusion != "" {
		matcher.Exclusion = []FileNamePattern{{Regex: config.Exclusion}}
		if config.Combine {
			matcher.Exclusion = append(matcher.Exclusion, DefaultMatchers.Exclusion...)
		}
	}

	return matcher
}

package testkeeper

import (
	"errors"

	"github.com/arquillian/ike-prow-plugins/pkg/scm"
)

// FileCategoryCounter is using plugin.FilePattern to figure out if the given commit affects any test file
// The plugin.FilePattern is loaded either from test-keeper.yaml file or from set of default matchers based on the languages using in the related project
type FileCategoryCounter struct {
	Matcher TestMatcher
}

// FileCategories holds information about the total files coming in the changeset, skipped files (those which are excluded from test verification)
// and tests
type FileCategories struct {
	Total, Skipped, Tests int
	Files                 *[]scm.ChangedFile
}

// OnlySkippedFiles indicates if changeset contains only files which are excluded from test verification
func (f *FileCategories) OnlySkippedFiles() bool {
	return f.Total > 0 && f.Skipped == f.Total
}

// TestsExist answers if any test files are found
func (f *FileCategories) TestsExist() bool {
	return f.Tests > 0
}

// NewFileTypes creates new instance of FileCategories struct with files populated
func NewFileTypes(files []scm.ChangedFile) FileCategories {
	return FileCategories{Files: &files, Total: len(files)}
}

// Count counts files in the changeset which are tests (included files) and should not be considered for
// verification (excluded). When first test is found it stops, as this is enough to unblock PR
func (t *FileCategoryCounter) Count(files []scm.ChangedFile) (FileCategories, error) {
	types := NewFileTypes(files)
	for _, file := range files {
		if file.Name == "" {
			return types, errors.New("can't have empty file name")
		}
		excluded := t.Matcher.MatchesExclusion(file.Name)
		if !excluded {
			if t.Matcher.MatchesInclusion(file.Name) {
				types.Tests++
				return types, nil // As we found the first test and we don't care about the amount of them, we can return
			}

		} else {
			types.Skipped++
		}
	}
	return types, nil
}

// LoadMatcher loads list of FilePattern either from the provided configuration or from languages retrieved from the given function
func LoadMatcher(configuration PluginConfiguration) (TestMatcher, error) {
	matcher, err := LoadDefaultMatcher()
	if err != nil {
		return matcher, err
	}

	if len(configuration.Inclusions) != 0 {
		exclusions := ParseFilePatterns(configuration.Inclusions)
		if configuration.Combine {
			matcher.Inclusion = append(matcher.Inclusion, exclusions...)
		} else {
			matcher.Inclusion = exclusions
		}
	}

	if len(configuration.Exclusions) != 0 {
		exclusions := ParseFilePatterns(configuration.Exclusions)
		if configuration.Combine {
			matcher.Exclusion = append(matcher.Exclusion, exclusions...)
		} else {
			matcher.Exclusion = exclusions
		}
	}

	return matcher, nil
}

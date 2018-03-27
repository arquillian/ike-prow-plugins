package plugin

import (
	"github.com/arquillian/ike-prow-plugins/pkg/assets"
	"github.com/arquillian/ike-prow-plugins/pkg/config"
	"github.com/pkg/errors"
)

// TestMatcher holds definitions of patterns considered as test filenames (inclusions) and those which shouldn't be
// verified (exclusions)
type TestMatcher struct {
	Inclusion []FilePattern
	Exclusion []FilePattern
}

// MatchesInclusion checks if file name matches defined inclusion patterns
func (matcher *TestMatcher) MatchesInclusion(filename string) bool {
	return Matches(matcher.Inclusion, filename)
}

// MatchesExclusion checks if file name matches defined exclusion patterns
func (matcher *TestMatcher) MatchesExclusion(filename string) bool {
	return Matches(matcher.Exclusion, filename)
}

// Matches iterates over a slice of FilePattern and verifies if passed name matches any of the defined patterns
func Matches(matchers []FilePattern, filename string) bool {

	for _, matcher := range matchers {
		if matcher.Matches(filename) {
			return true
		}
	}

	return false
}

// LoadDefaultMatcher loads default matcher containing default include and exclude patterns
func LoadDefaultMatcher() (TestMatcher, error) {
	matcher := TestMatcher{}
	defaultConfig := TestKeeperConfiguration{}

	err := config.Load(&defaultConfig, &assets.LocalLoadableConfig{ConfigFileName: "test-keeper.yaml"})

	if err != nil {
		return matcher, errors.Errorf("an error occurred while loading the default test-keeper.yaml: %s", err)
	}
	matcher.Inclusion = ParseFilePatterns(defaultConfig.Inclusions)
	matcher.Exclusion = ParseFilePatterns(defaultConfig.Exclusions)

	return matcher, nil
}

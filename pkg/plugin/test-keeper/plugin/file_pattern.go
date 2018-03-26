package plugin

import (
	"regexp"
	"strings"
)

const (
	regexDefinitionPrefix = "regex{{"
	regexDefinitionSuffix = "}}"
	anyPathWildcard       = "**"
	anyNameWildcard       = "*"
	anyNameRegex          = "[^/]*"
	anythingRegex         = ".*"
	twoStarsReplacement   = "<two-stars-replacement>"
	endOfLineRegex        = "$"
	directorySeparator    = "/"
)

// FilePattern contains regex that matches a file
type FilePattern struct {
	Regex string
}

// Matches checks if the given string (representing path to a file) contains a substring that matches Regex stored in this matcher
func (matcher *FilePattern) Matches(filename string) bool {
	return regexp.MustCompile(matcher.Regex).MatchString(filename)
}

// ParseFilePatterns takes the given patterns and parses to an array of FilePattern instances
func ParseFilePatterns(filePatterns []string) []FilePattern {
	patterns := make([]FilePattern, 0, len(filePatterns))
	for _, pattern := range filePatterns {
		patterns = append(patterns, FilePattern{
			Regex: parseFilePattern(strings.TrimSpace(pattern)),
		})
	}
	return patterns
}

func parseFilePattern(pattern string) string {

	// if it is regex{{...}} then just return the content
	if strings.HasPrefix(pattern, regexDefinitionPrefix) && strings.HasSuffix(pattern, regexDefinitionSuffix) {
		return pattern[len(regexDefinitionPrefix) : len(pattern)-len(regexDefinitionSuffix)]
	}

	// if not, then transform the pattern to regex
	slashIndex := strings.LastIndexAny(pattern, directorySeparator)

	path := transformPathPatternToRegex(pattern[:slashIndex+1])
	fileName := transformFilenamePatternToRegex(pattern[slashIndex+1:], path)

	regex := path + fileName

	if strings.HasSuffix(regex, directorySeparator) {
		regex = regex + anythingRegex
	} else {
		regex = regex + endOfLineRegex
	}

	return regex
}

func transformPathPatternToRegex(path string) string {
	path = escapeDots(path)
	path = strings.Replace(path, anyPathWildcard, twoStarsReplacement, -1)
	path = replaceAnyNameWildcards(path)
	return strings.Replace(path, twoStarsReplacement, anythingRegex, -1)
}

func transformFilenamePatternToRegex(fileName string, path string) string {
	fileName = escapeDots(fileName)

	if strings.HasPrefix(fileName, anyNameWildcard) {
		newPrefix := anyNameRegex
		if path == "" || len(fileName) == 1 {
			newPrefix = anythingRegex
		}
		return newPrefix + replaceAnyNameWildcards(fileName[1:])
	}
	return replaceAnyNameWildcards(fileName)

}

func escapeDots(s string) string {
	return strings.Replace(s, ".", "\\.", -1)
}

func replaceAnyNameWildcards(s string) string {
	return strings.Replace(s, anyNameWildcard, anyNameRegex, -1)
}

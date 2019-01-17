package prsanitizer

import (
	"strings"

	"regexp"

	"fmt"

	ghservice "github.com/arquillian/ike-prow-plugins/pkg/github/service"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	wip "github.com/arquillian/ike-prow-plugins/pkg/plugin/work-in-progress"
	gogh "github.com/google/go-github/github"
)

var (
	issueLinkRegexp = regexp.MustCompile(`(?i)(close|closes|closed|fix|fixes|fixed|resolve|resolves|resolved)[\s]*[:]?[\s]*[\w-\/]*#[\d]+`)
	defaultTypes    = []string{"chore", "docs", "feat", "fix", "refactor", "style", "test"}
)

const (
	// TitleFailureMessage is a message used in GH Status as description when the PR title does not follow semantic message style
	TitleFailureMessage = "#### Semantic title\nThe PR title `%s` does not conform with the [semantic message](https://seesparkbox.com/foundry/semantic_commit_messages) style. " +
		"The semantic message makes your changelog and git history clean. " +
		"Please, edit the PR by prefixing it with one of the type prefixes that are valid for your repository: %s."

	// DescriptionLengthShortMessage is a status message that is used in case of short PR description.
	DescriptionLengthShortMessage = "#### PR description length\nThe PR description is too short - it is expected that the description should have more than %d characters, but it has %d. " +
		"More elaborated description is helpful for understanding the changes proposed in this PR. (Any issue links and it's keywords are excluded when the description length is being measured)"

	// IssueLinkMissingMessage is a status message that is used in case of missing issue link.
	IssueLinkMissingMessage = "#### Issue link\nThe PR description is missing any issue link that would be used with any of the " +
		"[GitHub keywords](https://help.github.com/articles/closing-issues-using-keywords/). " +
		"Having it in the PR description ensures that the issue is automatically closed when the PR is merged."
)

type check func(pr *gogh.PullRequest, config PluginConfiguration, log log.Logger) string

func executeChecks(pr *gogh.PullRequest, config PluginConfiguration, log log.Logger) []string {
	checks := []check{CheckSemanticTitle, CheckDescriptionLength, CheckIssueLinkPresence}
	var messages []string
	for _, check := range checks {
		msg := check(pr, config, log)
		if msg != "" {
			messages = append(messages, msg)
		}
	}
	return messages
}

// CheckSemanticTitle checks if the given PR contains semantic title
func CheckSemanticTitle(pr *gogh.PullRequest, config PluginConfiguration, log log.Logger) string {
	change := ghservice.NewRepositoryChangeForPR(pr)
	prefixes := GetValidTitlePrefixes(config)
	isTitleWithValidType := HasTitleWithValidType(prefixes, *pr.Title)

	if !isTitleWithValidType {
		if prefix, ok := wip.GetWorkInProgressPrefix(*pr.Title, wip.LoadConfiguration(log, change)); ok {
			trimmedTitle := strings.TrimPrefix(*pr.Title, prefix)
			isTitleWithValidType = HasTitleWithValidType(prefixes, trimmedTitle)
		}
	}
	if !isTitleWithValidType {
		allPrefixes := "`" + strings.Join(prefixes, "`, `") + "`"
		return fmt.Sprintf(TitleFailureMessage, pr.GetTitle(), allPrefixes)
	}
	return ""
}

// CheckDescriptionLength  checks if the given PR's description contains enough number of arguments
func CheckDescriptionLength(pr *gogh.PullRequest, config PluginConfiguration, log log.Logger) string {
	actualLength := len(strings.TrimSpace(issueLinkRegexp.ReplaceAllString(pr.GetBody(), "")))
	if actualLength < config.DescriptionContentLength {
		return fmt.Sprintf(DescriptionLengthShortMessage, config.DescriptionContentLength, actualLength)
	}
	return ""
}

// CheckIssueLinkPresence checks if the given PR's description contains an issue link
func CheckIssueLinkPresence(pr *gogh.PullRequest, config PluginConfiguration, log log.Logger) string {
	if !issueLinkRegexp.MatchString(pr.GetBody()) {
		return IssueLinkMissingMessage
	}
	return ""
}

// GetValidTitlePrefixes returns list of valid prefixes
func GetValidTitlePrefixes(config PluginConfiguration) []string {
	prefixes := defaultTypes
	if len(config.TypePrefix) != 0 {
		if config.Combine {
			prefixes = append(prefixes, config.TypePrefix...)
		} else {
			prefixes = config.TypePrefix
		}
	}
	return prefixes
}

// HasTitleWithValidType checks if title prefix conforms with semantic message style.
func HasTitleWithValidType(prefixes []string, title string) bool {
	pureTitle := strings.TrimSpace(title)
	for _, prefix := range prefixes {
		prefixRegexp := regexp.MustCompile(`(?i)^` + prefix + `(:| |\()+`)
		if prefixRegexp.MatchString(pureTitle) {
			return true
		}
	}
	return false
}

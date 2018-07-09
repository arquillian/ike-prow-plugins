package prsanitizer

import (
	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/status"
)

type prSanitizerStatusService struct {
	statusService scm.StatusService
}

const (
	// ProwPluginName is an external prow plugin name used to register this service
	ProwPluginName = "pr-sanitizer"

	// FailureDetailsPageName is a name of a documentation page that contains additional status details for title verification failure.
	FailureDetailsPageName = "pr-sanitizer-failed"

	// FailureMessage is a message used in GH Status as description when the PR title and description does not conform to the PR sanitizer checks.
	FailureMessage = "Meh! Some PR Sanitizer Standard Check failed. :("

	// SuccessMessage is a message used in GH Status as description when the PR title and description conforms to the PR sanitizer checks.
	SuccessMessage = "Yay! All PR Sanitizer title and description checks passed. :)"

	// SuccessDetailsPageName is a name of a documentation page that contains additional status details for success state
	SuccessDetailsPageName = "pr-sanitizer-success"
)

func (gh *GitHubPRSanitizerEventsHandler) newPRTitleDescriptionStatusService(log log.Logger, change scm.RepositoryChange) prSanitizerStatusService {
	statusContext := github.StatusContext{BotName: gh.BotName, PluginName: ProwPluginName}
	statusService := status.NewStatusService(gh.Client, log, change, statusContext)
	return prSanitizerStatusService{statusService: statusService}
}

func (ss *prSanitizerStatusService) titleAndDescriptionOk() error {
	return ss.statusService.Success(SuccessMessage, SuccessDetailsPageName)
}

func (ss *prSanitizerStatusService) fail() error {
	return ss.statusService.Failure(FailureMessage, FailureDetailsPageName)
}

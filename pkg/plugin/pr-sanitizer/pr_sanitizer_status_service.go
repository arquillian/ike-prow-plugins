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

	// SuccessMessage is a message used in GH Status as description when the PR title conforms to the semantic commit message style
	SuccessMessage = "PR title conforms with semantic commit message style and description has enough characters with issue link."

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

func (ss *prSanitizerStatusService) fail(fm FailureMessage) error {
	return ss.statusService.Failure(string(fm), FailureDetailsPageName)
}

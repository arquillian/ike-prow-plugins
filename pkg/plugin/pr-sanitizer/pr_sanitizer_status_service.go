package prsanitizer

import (
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/status"
	"github.com/arquillian/ike-prow-plugins/pkg/status/message"
	gogh "github.com/google/go-github/github"
)

type prSanitizerStatusService struct {
	statusService    scm.StatusService
	statusMsgService *message.StatusMessageService
	log              log.Logger
}

const (
	// ProwPluginName is an external prow plugin name used to register this service
	ProwPluginName = "pr-sanitizer"

	// FailureDetailsPageName is a name of a documentation page that contains additional status details for title verification failure.
	FailureDetailsPageName = "pr-sanitizer-failed"

	// FailureMessage is a message used in GH Status as description when the PR title and description does not conform to the PR sanitizer checks.
	FailureMessage = "This PR doesn't comply with PR conventions :("

	// SuccessMessage is a message used in GH Status as description when the PR title and description conforms to the PR sanitizer checks.
	SuccessMessage = "This PR complies with PR conventions :)"

	// SuccessDetailsPageName is a name of a documentation page that contains additional status details for success state
	SuccessDetailsPageName = "pr-sanitizer-success"

	// FailureStatusMessageBeginning is a beginning of failure status message
	FailureStatusMessageBeginning = "This pull request doesn't comply with the conventions given by the `pr-sanitizer` plugin. " +
		"The following items are necessary to fix:\n\n"

	// SuccessStatusMessage is a status message used when the PR is good
	SuccessStatusMessage = "This pull request complies with the PR conventions given by the `pr-sanitizer` plugin. :)"
)

func (gh *GitHubPRSanitizerEventsHandler) newPrSanitizerStatusService(log log.Logger, pr *gogh.PullRequest, config PluginConfiguration) prSanitizerStatusService {
	statusContext := github.StatusContext{BotName: gh.BotName, PluginName: ProwPluginName}

	change := ghservice.NewRepositoryChangeForPR(pr)
	statusService := status.NewStatusService(gh.Client, log, change, statusContext)

	commentsLoader := ghservice.NewIssueCommentsLazyLoader(gh.Client, pr)
	msgContext := message.NewStatusMessageContext(ProwPluginName, documentationSection, pr, config.PluginConfiguration)
	msgService := message.NewStatusMessageService(gh.Client, log, commentsLoader, msgContext)

	return prSanitizerStatusService{
		statusService:    statusService,
		statusMsgService: msgService,
		log:              log,
	}
}

func (ss *prSanitizerStatusService) success() error {
	ss.statusMsgService.HappyStatusMessage(SuccessStatusMessage, "success", false)
	return ss.statusService.Success(SuccessMessage, SuccessDetailsPageName)
}

func (ss *prSanitizerStatusService) fail(messages []string) error {
	msg := FailureStatusMessageBeginning + strings.Join(messages, "\n\n")
	ss.statusMsgService.SadStatusMessage(msg, "failed", true)
	return ss.statusService.Failure(FailureMessage, FailureDetailsPageName)
}

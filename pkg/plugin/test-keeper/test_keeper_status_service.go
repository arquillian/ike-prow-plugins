package testkeeper

import (
	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/status"
	"github.com/arquillian/ike-prow-plugins/pkg/status/message"
	gogh "github.com/google/go-github/github"
)

type testStatusService struct {
	log           log.Logger
	change        scm.RepositoryChange
	statusService scm.StatusService
}

const (
	// TestsExistMessage is a message used in GH Status as description when tests are found
	TestsExistMessage = "There are some tests :)"
	// TestsExistDetailsPageName is a name of a documentation page that contains additional status details for TestsExistMessage
	TestsExistDetailsPageName = "tests-exist"

	// NoTestsMessage is a message used in GH Status as description when no tests shipped with the PR
	NoTestsMessage = "No tests in this PR :("
	// NoTestsDetailsPageName is a name of a documentation page that contains additional status details for NoTestsMessage
	NoTestsDetailsPageName = "no-tests"

	// OkOnlySkippedFilesMessage is a message used in GH Status as description when PR comes with a changeset which shouldn't be subject of test verification
	OkOnlySkippedFilesMessage = "Seems that this PR doesn't need to have tests"
	// OkOnlySkippedFilesDetailsPageName is a name of a documentation page that contains additional status details for OkOnlySkippedFilesMessage
	OkOnlySkippedFilesDetailsPageName = "only-skipped"

	// FailureMessage is a message used in GH Status as description when failure occurred
	FailureMessage = "Failed while check for tests"

	// ApprovedByMessage is a message used in GH Status as description when it's commented to skip the check
	ApprovedByMessage = "PR is fine without tests says @%s"
	// ApprovedByDetailsPageName is a name of a documentation page that contains additional status details for ApprovedByMessage
	ApprovedByDetailsPageName = "keeper-approved-by"
)

func (gh *GitHubTestEventsHandler) newTestStatusService(log log.Logger, pullRequest *gogh.PullRequest) *testStatusService {
	change := ghservice.NewRepositoryChangeForPR(pullRequest)
	statusContext := github.StatusContext{BotName: gh.BotName, PluginName: ProwPluginName}
	statusService := status.NewStatusService(gh.Client, log, change, statusContext)
	return &testStatusService{
		log:           log,
		change:        change,
		statusService: statusService,
	}
}

func (ts *testStatusService) okTestsExist() error {
	return ts.statusService.Success(TestsExistMessage, TestsExistDetailsPageName)
}

func (ts *testStatusService) okOnlySkippedFiles() error {
	return ts.statusService.Success(OkOnlySkippedFilesMessage, OkOnlySkippedFilesDetailsPageName)
}

func (ts *testStatusService) okWithoutTests(approvedBy string) error {
	return ts.statusService.Success(fmt.Sprintf(ApprovedByMessage, approvedBy), ApprovedByDetailsPageName)
}

func (ts *testStatusService) reportError() error {
	return ts.statusService.Error(FailureMessage)
}

func (ts *testStatusService) failNoTests() error {
	return ts.statusService.Failure(NoTestsMessage, NoTestsDetailsPageName)
}

const (
	paragraph = "\n\n"

	// WithoutTestsMsg contains a status message related to the state when PR is pushed without any test
	WithoutTestsMsg = "It appears that no tests have been added or updated in this PR." +
		paragraph +
		"Automated tests give us confidence in shipping reliable software. Please add some as part of this change." +
		paragraph +
		"If you are an admin or the reviewer of this PR and you are sure that no test is needed then you can use the command `" + BypassCheckComment + "` " +
		"as a comment to make the status green.\n"

	documentationSection = "#_test_keeper_plugin"

	// WithTestsMsg contains a status message related to the state when PR is updated by a commit containing a test
	WithTestsMsg = "It seems that this PR already contains some added or changed tests. Good job!"

	// OnlySkippedMsg contains a status message related to the state when PR is updated so it contains only skipped files
	OnlySkippedMsg = "It seems that this PR doesn't need any test as all changed files in the changeset match " +
		"patterns for which the validation should be skipped."
)

type testStatusServiceWithMessages struct {
	*testStatusService
	statusMsgService *message.StatusMessageService
	config           PluginConfiguration
}

func (gh *GitHubTestEventsHandler) newTestStatusServiceWithMessages(log log.Logger, pullRequest *gogh.PullRequest,
	commentsLoader *ghservice.IssueCommentsLazyLoader, config PluginConfiguration) testStatusServiceWithMessages {

	msgContext := message.NewStatusMessageContext(ProwPluginName, documentationSection, pullRequest, config.PluginConfiguration)
	msgService := message.NewStatusMessageService(gh.Client, log, commentsLoader, msgContext)

	return testStatusServiceWithMessages{
		testStatusService: gh.newTestStatusService(log, pullRequest),
		statusMsgService:  msgService,
		config:            config,
	}
}

// CreateWithoutTestsMessage creates a status message for the test-keeper plugin. If the status message is set in config then it takes that one, the default otherwise.
func (ts *testStatusServiceWithMessages) withoutTestsMessage() {
	ts.statusMsgService.SadStatusMessage(WithoutTestsMsg, "without_tests", true)
}

// CreateWithTestsMessage creates a status message for the test-keeper plugin. If the status message is set in config then it takes that one, the default otherwise.
func (ts *testStatusServiceWithMessages) withTestsMessage() {
	ts.statusMsgService.HappyStatusMessage(WithTestsMsg, "with_tests", false)
}

// CreateOnlySkippedMessage creates a status message for the test-keeper plugin. If the status message is set in config then it takes that one, the default otherwise.
func (ts *testStatusServiceWithMessages) onlySkippedMessage() {
	ts.statusMsgService.HappyStatusMessage(OnlySkippedMsg, "only_skipped", false)
}

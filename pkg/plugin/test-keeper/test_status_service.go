package testkeeper

import (
	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
)

type testStatusService struct {
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

func (gh *GitHubTestEventsHandler) newTestStatusService(log log.Logger, change scm.RepositoryChange) testStatusService {
	statusContext := github.StatusContext{BotName: gh.BotName, PluginName: ProwPluginName}
	statusService := ghservice.NewStatusService(gh.Client, log, change, statusContext)
	return testStatusService{statusService: statusService}
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

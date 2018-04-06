package testkeeper

import (
	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
)

type testStatusService struct {
	statusService scm.StatusService
}

const (
	// TestsExistMessage is a message used in GH Status as description when tests are found
	TestsExistMessage = "There are some tests :)"
	// TestsExistDetailsLink is a link to an anchor in arq documentation that contains additional status details for TestsExistMessage
	TestsExistDetailsLink = plugin.DocumentationURL + "#tests-exist"

	// NoTestsMessage is a message used in GH Status as description when no tests shipped with the PR
	NoTestsMessage = "No tests in this PR :("
	// NoTestsDetailsLink is a link to an anchor in arq documentation that contains additional status details for NoTestsMessage
	NoTestsDetailsLink = plugin.DocumentationURL + "#no-tests"

	// OkOnlySkippedFilesMessage is a message used in GH Status as description when PR comes with a changeset which shouldn't be subject of test verification
	OkOnlySkippedFilesMessage = "Seems that this PR doesn't need to have tests"
	// OkOnlySkippedFilesDetailsLink is a link to an anchor in arq documentation that contains additional status details for OkOnlySkippedFilesMessage
	OkOnlySkippedFilesDetailsLink = plugin.DocumentationURL + "#only-skipped"

	// FailureMessage is a message used in GH Status as description when failure occured
	FailureMessage = "Failed while check for tests"

	// ApprovedByMessage is a message used in GH Status as description when it's commented to skip the check
	ApprovedByMessage = "PR is fine without tests says @%s"
	// ApprovedByDetailsLink is a link to an anchor in arq documentation that contains additional status details for ApprovedByMessage
	ApprovedByDetailsLink = plugin.DocumentationURL + "#keeper-approved-by"
)

func (gh *GitHubTestEventsHandler) newTestStatusService(log log.Logger, change scm.RepositoryChange) testStatusService {
	statusContext := github.StatusContext{BotName: "alien-ike", PluginName: ProwPluginName}
	statusService := github.NewStatusService(gh.Client, log, change, statusContext)
	return testStatusService{statusService: statusService}
}

func (ts *testStatusService) okTestsExist() error {
	return ts.statusService.Success(TestsExistMessage, TestsExistDetailsLink)
}

func (ts *testStatusService) okOnlySkippedFiles() error {
	return ts.statusService.Success(OkOnlySkippedFilesMessage, OkOnlySkippedFilesDetailsLink)
}

func (ts *testStatusService) okWithoutTests(approvedBy string) error {
	return ts.statusService.Success(fmt.Sprintf(ApprovedByMessage, approvedBy), ApprovedByDetailsLink)
}

func (ts *testStatusService) reportError() error {
	return ts.statusService.Error(FailureMessage)
}

func (ts *testStatusService) failNoTests() error {
	return ts.statusService.Failure(NoTestsMessage, NoTestsDetailsLink)
}

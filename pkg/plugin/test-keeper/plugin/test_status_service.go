package plugin

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
	// TestsExistTargetURL is a link to an anchor in arq documentation that contains additional status details for TestsExistMessage
	TestsExistTargetURL = plugin.DocumentationURL + "#tests-exist"

	// NoTestsMessage is a message used in GH Status as description when no tests shipped with the PR
	NoTestsMessage = "No tests in this PR :("
	// NoTestsTargetURL is a link to an anchor in arq documentation that contains additional status details for NoTestsMessage
	NoTestsTargetURL = plugin.DocumentationURL + "#no-tests"

	// OkOnlySkippedFilesMessage is a message used in GH Status as description when PR comes with a changeset which shouldn't be subject of test verification
	OkOnlySkippedFilesMessage = "Seems that this PR doesn't need to have tests"
	// OkOnlySkippedFilesTargetURL is a link to an anchor in arq documentation that contains additional status details for OkOnlySkippedFilesMessage
	OkOnlySkippedFilesTargetURL = plugin.DocumentationURL + "#only-skipped"

	// FailureMessage is a message used in GH Status as description when failure occured
	FailureMessage = "Failed while check for tests"

	// ApprovedByMessage is a message used in GH Status as description when it's commented to skip the check
	ApprovedByMessage = "PR is fine without tests says @%s"
	// ApprovedByTargetURL is a link to an anchor in arq documentation that contains additional status details for ApprovedByMessage
	ApprovedByTargetURL = plugin.DocumentationURL + "#keeper-approved-by"
)

func (gh *GitHubTestEventsHandler) newTestStatusService(log log.Logger, change scm.RepositoryChange) testStatusService {
	statusContext := github.StatusContext{BotName: "alien-ike", PluginName: ProwPluginName}
	statusService := github.NewStatusService(gh.Client, log, change, statusContext)
	return testStatusService{statusService: statusService}
}

func (ts *testStatusService) okTestsExist() error {
	return ts.statusService.Success(TestsExistMessage, TestsExistTargetURL)
}

func (ts *testStatusService) okOnlySkippedFiles() error {
	return ts.statusService.Success(OkOnlySkippedFilesMessage, OkOnlySkippedFilesTargetURL)
}

func (ts *testStatusService) okWithoutTests(approvedBy string) error {
	return ts.statusService.Success(fmt.Sprintf(ApprovedByMessage, approvedBy), ApprovedByTargetURL)
}

func (ts *testStatusService) reportError() error {
	return ts.statusService.Error(FailureMessage, "") // TODO create link to detailed log about the problem
}

func (ts *testStatusService) failNoTests() error {
	return ts.statusService.Failure(NoTestsMessage, NoTestsTargetURL)
}

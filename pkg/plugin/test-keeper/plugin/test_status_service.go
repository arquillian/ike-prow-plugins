package plugin

import (
	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
)

type testStatusService struct {
	statusService scm.StatusService
}

const (
	// TestsExistMessage is a message used in GH Status as description when tests are found
	TestsExistMessage = "There are some tests :)"
	// NoTestsMessage is a message used in GH Status as description when no tests shipped with the PR
	NoTestsMessage = "No tests in this PR :("
	// OkOnlySkippedFilesMessage is a message used in GH Status as description when PR comes with a changeset which shouldn't be subject of test verification
	OkOnlySkippedFilesMessage = "Seems that this PR doesn't need to have tests"
	// FailureMessage is a message used in GH Status as description when failure occured
	FailureMessage = "Failed while check for tests"
	// ApproveByMessage is a message used in GH Status as description when it's commented to skip the check
	ApproveByMessage = "PR is fine without tests says @%s"
)

func (gh *GitHubTestEventsHandler) newTestStatusService(log log.Logger, change scm.RepositoryChange) testStatusService {
	statusContext := github.StatusContext{BotName: "alien-ike", PluginName: ProwPluginName}
	statusService := github.NewStatusService(gh.Client, log, change, statusContext)
	return testStatusService{statusService: statusService}
}

func (ts *testStatusService) okTestsExist() error {
	return ts.statusService.Success(TestsExistMessage)
}

func (ts *testStatusService) okOnlySkippedFiles() error {
	return ts.statusService.Success(OkOnlySkippedFilesMessage) // TODO create link to detailed log about the problem
}

func (ts *testStatusService) okWithoutTests(approvedBy string) error {
	return ts.statusService.Success(fmt.Sprintf(ApproveByMessage, approvedBy))
}

func (ts *testStatusService) reportError() error {
	return ts.statusService.Error(FailureMessage) // TODO create link to detailed log about the problem
}

func (ts *testStatusService) failNoTests() error {
	return ts.statusService.Failure(NoTestsMessage)
}

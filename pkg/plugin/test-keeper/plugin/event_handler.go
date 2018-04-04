package plugin

import (
	"encoding/json"
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	gogh "github.com/google/go-github/github"
)

// GitHubTestEventsHandler is the event handler for the plugin.
// Implements server.GitHubEventHandler interface which contains the logic for incoming GitHub events
type GitHubTestEventsHandler struct {
	Client *github.Client
}

const (
	// ProwPluginName is an external prow plugin name used to register this service
	ProwPluginName = "test-keeper"
	// SkipComment is used as a command to bypass test presence validation
	SkipComment    = "/ok-without-tests"
)

var (
	handledPrActions      = []string{"opened", "reopened", "edited", "synchronize"}
	handledCommentActions = []string{"created", "deleted"}
)

// HandleEvent is an entry point for the plugin logic. This method is invoked by the Server when
// events are dispatched from the /hook service
func (gh *GitHubTestEventsHandler) HandleEvent(log log.Logger, eventType github.EventType, payload []byte) error {

	switch eventType {
	case github.PullRequest:
		var event gogh.PullRequestEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			log.Errorf("Failed while parsing '%q' event with payload: %q. Cause: %q", github.PullRequest, event, err)
			return err
		}

		if err := gh.handlePrEvent(log, &event); err != nil {
			log.Errorf("Error handling '%q' event with payload %q. Cause: %q", github.PullRequest, event, err)
			return err
		}

	case github.IssueComment:
		var event gogh.IssueCommentEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			log.Errorf("Failed while parsing '%q' event with payload: %q. Cause: %q", github.IssueComment, event, err)
			return err
		}

		if err := gh.handlePrComment(log, &event); err != nil {
			log.Errorf("Error handling '%q' event with payload %q. Cause: %q", github.IssueComment, event, err)
			return err
		}

	default:
		log.Warnf("received an event of type %q but didn't ask for it", eventType)
	}
	return nil
}

func (gh *GitHubTestEventsHandler) handlePrEvent(log log.Logger, event *gogh.PullRequestEvent) error {
	if !utils.Contains(handledPrActions, *event.Action) {
		return nil
	}

	change := scm.RepositoryChange{
		Owner:    *event.Repo.Owner.Login,
		RepoName: *event.Repo.Name,
		Hash:     *event.PullRequest.Head.SHA,
	}

	return gh.checkTestsAndSetStatus(log, change, event.PullRequest)
}

func (gh *GitHubTestEventsHandler) handlePrComment(log log.Logger, prComment *gogh.IssueCommentEvent) error {
	if !utils.Contains(handledCommentActions, *prComment.Action) {
		return nil
	}

	org := prComment.Repo.Owner.Login
	name := prComment.Repo.Name
	prNumber := prComment.Issue.Number

	sender := prComment.Sender.Login
	permissionLevel, e := gh.Client.GetPermissionLevel(*org, *name, *sender)
	if e != nil {
		log.Fatal(e)
		return e
	}

	if *permissionLevel.Permission != "admin" {
		return nil
	}

	comment := strings.TrimSpace(*prComment.Comment.Body)

	if comment != SkipComment {
		return nil
	}

	pullRequest, err := gh.Client.GetPullRequest(*org, *name, *prNumber)
	if err != nil {
		log.Fatal(err)
		return err
	}

	change := scm.RepositoryChange{
		Owner:    *prComment.Repo.Owner.Login,
		RepoName: *prComment.Repo.Name,
		Hash:     *pullRequest.Head.SHA,
	}

	if comment == SkipComment && *prComment.Action == "deleted" {
		return gh.checkTestsAndSetStatus(log, change, pullRequest)
	}

	statusService := gh.newTestStatusService(log, change)
	return statusService.okWithoutTests(*sender)
}

func (gh *GitHubTestEventsHandler) checkTestsAndSetStatus(log log.Logger, change scm.RepositoryChange, pr *gogh.PullRequest) error {
	statusService := gh.newTestStatusService(log, change)
	configuration := LoadTestKeeperConfig(log, change)
	fileCategories, err := gh.checkTests(log, change, configuration, *pr.Number)
	if err != nil {
		if statusErr := statusService.reportError(); statusErr != nil {
			log.Errorf("failed to report error status on PR [%q]. cause: %s", *pr, statusErr)
		}
		return err
	}

	if fileCategories.OnlySkippedFiles() {
		return statusService.okOnlySkippedFiles()
	}

	if fileCategories.TestsExist() {
		return statusService.okTestsExist()
	}

	err = statusService.failNoTests()
	if err != nil {
		log.Errorf("failed to report status on PR [%q]. cause: %s", *pr, err)
	}

	hintContext := github.HintContext{PluginName: ProwPluginName, Assignee: *pr.User.Login}
	hinter := github.NewHinter(gh.Client, log, change, *pr.Number, hintContext)

	cerr := hinter.PluginComment(CreateCommentMessage(configuration, change))
	if cerr != nil {
		log.Errorf("failed to comment on PR [%q]. cause: %s", *pr, cerr)
		return cerr
	}

	return err
}

func (gh *GitHubTestEventsHandler) checkTests(log log.Logger, change scm.RepositoryChange, config TestKeeperConfiguration, prNumber int) (FileCategories, error) {
	matcher, err := LoadMatcher(config)
	if err != nil {
		log.Error(err)
		return FileCategories{}, err
	}

	fileCategoryCounter := FileCategoryCounter{Matcher: matcher}

	changedFiles, err := gh.Client.ListPullRequestFiles(change.Owner, change.RepoName, prNumber)
	if err != nil {
		log.Error(err)
		return FileCategories{}, err
	}

	fileCategories, err := fileCategoryCounter.Count(changedFiles)
	if err != nil {
		log.Error(err)
	}

	return fileCategories, err
}

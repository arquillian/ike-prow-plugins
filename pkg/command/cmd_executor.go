package command

import (
	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	gogh "github.com/google/go-github/github"
	"strings"
)

// DoFunction is used for performing operations related to command actions
type DoFunction func() error
type doFunctionExecutor func(client ghclient.Client, log log.Logger, comment *gogh.IssueCommentEvent) error

// CmdExecutor takes care of executing a command triggered by IssueCommentEvent.
// The execution is set by specifying actions/events and with given restrictions the command should be triggered for.
type CmdExecutor struct {
	Command   string
	Quiet     bool
	executors []doFunctionExecutor
}

// RestrictionSetter keeps information about set actions the command should be triggered for and opens an API to provide
// permission restrictions
type RestrictionSetter struct {
	commandExecutor *CmdExecutor
	actions         []commentAction
}

// DoFunctionProvider keeps all allowed actions and permission checks and opens an API to provide a DoFunction implementation
type DoFunctionProvider struct {
	commandExecutor  *CmdExecutor
	actions          []commentAction
	permissionChecks []PermissionCheck
}

type commentAction struct {
	actions     []string
	description string
	log         bool
}

// Deleted represents comment deletion
var Deleted = commentAction{actions: []string{"deleted"}, description: "delete", log: false}

// Triggered represents comment editions and creation
var Triggered = commentAction{actions: []string{"edited", "created"}, description: "trigger", log: true}

func (a *commentAction) isMatching(comment *gogh.IssueCommentEvent) bool {
	return utils.Contains(a.actions, *comment.Action)
}

// When takes list of actions the command should be triggered for
func (e *CmdExecutor) When(actions ...commentAction) *RestrictionSetter {
	return &RestrictionSetter{commandExecutor: e, actions: actions}
}

// By takes a list of permission checks the command should be restricted by
func (s *RestrictionSetter) By(permissionChecks ...PermissionCheck) *DoFunctionProvider {
	return &DoFunctionProvider{commandExecutor: s.commandExecutor, actions: s.actions, permissionChecks: permissionChecks}
}

// Then take a DoFunction that performs the required operations (when all checks are fulfilled)
func (p *DoFunctionProvider) Then(doFunction DoFunction) {
	doExecutor := func(client ghclient.Client, log log.Logger, comment *gogh.IssueCommentEvent) error {
		matchingAction := p.getMatchingAction(comment)
		if matchingAction == nil {
			return nil
		}

		status, err := AllOf(p.permissionChecks...)(true)
		if status.UserIsApproved && err == nil {
			return doFunction()
		}
		message := status.constructMessage(matchingAction.description, p.commandExecutor.Command)
		log.Warn(message)
		if err == nil && matchingAction.log && !p.commandExecutor.Quiet {
			commentService := ghservice.NewCommentService(client, comment)
			return commentService.AddComment(&message)
		}
		return err
	}

	p.commandExecutor.executors = append(p.commandExecutor.executors, doExecutor)
}

func (p *DoFunctionProvider) getMatchingAction(comment *gogh.IssueCommentEvent) *commentAction {
	for _, action := range p.actions {
		if action.isMatching(comment) {
			return &action
		}
	}
	return nil
}

// Execute triggers the given DoFunctions (when all checks are fulfilled) for the given pr comment
func (e *CmdExecutor) Execute(client ghclient.Client, log log.Logger, comment *gogh.IssueCommentEvent) error {
	body := strings.TrimSpace(*comment.Comment.Body)
	if prefix := strings.Split(body, " ")[0]; e.Command != body && prefix != e.Command {
		return nil
	}
	for _, doExecutor := range e.executors {
		err := doExecutor(client, log, comment)
		if err != nil {
			return err
		}
	}
	return nil
}

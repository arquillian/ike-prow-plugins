package command

import (
	"strings"

	ghclient "github.com/arquillian/ike-prow-plugins/pkg/github/client"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	gogh "github.com/google/go-github/v41/github"
)

// RunCommentPrefix is used as a command prefix to trigger plugin with it's name
const RunCommentPrefix = "/run"

// RunCmd represents a command that is triggered by "/run plugin-name" or "/run all"
type RunCmd struct {
	PluginName            string
	UserPermissionService *PermissionService
	WhenAddedOrEdited     DoFunction
}

// Perform executes the set DoFunctions for the given IssueCommentEvent (when all conditions are fulfilled)
func (c *RunCmd) Perform(client ghclient.Client, logger log.Logger, comment *gogh.IssueCommentEvent) error {
	user := c.UserPermissionService
	var RunCommand = &CmdExecutor{Command: RunCommentPrefix}

	RunCommand.
		When(Triggered).
		By(AnyOf(user.Admin, user.PRReviewer, user.PRApprover, user.PRCreator)).
		Then(c.WhenAddedOrEdited)

	return RunCommand.Execute(client, logger, comment)
}

// Matches returns true when the given IssueCommentEvent content prefix is "/run"
func (c *RunCmd) Matches(comment *gogh.IssueCommentEvent) bool {
	body := strings.TrimSpace(*comment.Comment.Body)
	command := strings.Split(body, " ")
	pluginNames := command[1:]
	return command[0] == RunCommentPrefix && (utils.Contains(pluginNames, c.PluginName) || utils.Contains(pluginNames, "all"))
}

package plugin

import (
	"strings"

	is "github.com/arquillian/ike-prow-plugins/pkg/command"
	"github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	gogh "github.com/google/go-github/github"
)

// SkipComment is used as a command to bypass test presence validation
const SkipComment = "/ok-without-tests"

// SkipCommentCmd represents a command that is triggered by "/ok-without-tests"
type SkipCommentCmd struct {
	userPermissionService *is.PermissionService
	whenDeleted           is.DoFunction
	whenAddedOrCreated    is.DoFunction
}

// Perform executes the set DoFunctions for the given IssueCommentEvent (when all conditions are fulfilled)
func (c *SkipCommentCmd) Perform(client *github.Client, log log.Logger, comment *gogh.IssueCommentEvent) error {
	user := c.userPermissionService
	var SkipCommentCommand = &is.CmdExecutor{Command: SkipComment}

	SkipCommentCommand.When(is.Deleted).By(is.Anybody).Then(c.whenDeleted)

	SkipCommentCommand.
		When(is.Triggered).
		By(is.AnyOf(user.Admin, user.PRReviewer), is.Not(user.PRCreator)).
		Then(c.whenAddedOrCreated)

	return SkipCommentCommand.Execute(client, log, comment)
}

// Matches returns true when the given IssueCommentEvent content is same as "/ok-without-tests"
func (c *SkipCommentCmd) Matches(comment *gogh.IssueCommentEvent) bool {
	body := strings.TrimSpace(*comment.Comment.Body)
	return body == SkipComment
}

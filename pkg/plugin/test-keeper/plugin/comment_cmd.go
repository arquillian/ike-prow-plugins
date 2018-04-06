package plugin

import (
	"strings"

	is "github.com/arquillian/ike-prow-plugins/pkg/command"
	"github.com/arquillian/ike-prow-plugins/pkg/github"
	gogh "github.com/google/go-github/github"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
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
func (c *SkipCommentCmd) Perform(client *github.Client, log log.Logger, prComment *gogh.IssueCommentEvent) error {
	user := c.userPermissionService
	var SkipCommentCommand = &is.CmdExecutor{Command: SkipComment}

	SkipCommentCommand.When(is.Deleted).By(is.AnyBody).ThenDo(c.whenDeleted)

	SkipCommentCommand.
		When(is.Triggered).
		By(is.AnyOf(user.ThatIsAdmin, user.ThatIsPRReviewer), is.Not(user.ThatIsPRCreator)).
		ThenDo(c.whenAddedOrCreated)

	return SkipCommentCommand.Execute(client, log, prComment)
}

// Matches returns true when the given IssueCommentEvent content is same as "/ok-without-tests"
func (c *SkipCommentCmd) Matches(prComment *gogh.IssueCommentEvent) bool {
	comment := strings.TrimSpace(*prComment.Comment.Body)
	return comment == SkipComment
}

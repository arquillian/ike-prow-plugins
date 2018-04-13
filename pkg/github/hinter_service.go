package github

import (
	"fmt"
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
)

// PluginTitleTemplate is a constant template containing "Ike Plugins (name-of-plugin)" title with markdown formatting
const (
	PluginTitleTemplate     = "### Ike Plugins (%s)"
	assigneeMentionTemplate = "Thank you @%s for this contribution!"
)

// Hinter is a struct managing plugin comments
type Hinter struct {
	*CommentService
	log            log.Logger
	commentContext HintContext
}

// HintContext holds a plugin name and a assignee to be mentioned in the comment
type HintContext struct {
	PluginName string
	Assignee   string // TODO rethink this naming when plugins will start interacting with issue creators and reviewers
}

// NewHinter creates an instance of GitHub Hinter for the given HintContext
func NewHinter(client Client, log log.Logger, change scm.RepositoryChange, issueOrPrNumber int, commentContext HintContext) *Hinter {
	return &Hinter{
		CommentService: &CommentService{
			client: client,
			issue: scm.RepositoryIssue{
				Owner:    change.Owner,
				RepoName: change.RepoName,
				Number:   issueOrPrNumber,
			},
		},
		log:            log,
		commentContext: commentContext,
	}
}

// PluginComment checks all present comments in the issue/pull-request. If no comment with PluginTitleTemplate
// (with the related plugin) is found, then it adds a new comment with the plugin title, assignee mention
// and the given commentMsg. If such a comment is present already, then it does nothing.
func (s *Hinter) PluginComment(commentMsg string) error {

	comments, err := s.client.ListIssueComments(s.issue)
	if err != nil {
		s.log.Errorf("Getting all comments failed with an error: %s", err)
	}

	for _, com := range comments {
		content := *com.Body
		if strings.HasPrefix(content, s.getPluginTitle()) {
			return nil
		}
	}

	return s.AddComment(s.createPluginHint(commentMsg))
}

func (s *Hinter) append(first, second string) string {
	return first + "\n\n" + second
}

func (s *Hinter) createPluginHint(commentMsg string) *string {
	return utils.String(s.append(s.createBeginning(), commentMsg))
}

func (s *Hinter) createBeginning() string {
	return s.append(s.getPluginTitle(), fmt.Sprintf(assigneeMentionTemplate, s.commentContext.Assignee))
}

func (s *Hinter) getPluginTitle() string {
	return fmt.Sprintf(PluginTitleTemplate, s.commentContext.PluginName)
}

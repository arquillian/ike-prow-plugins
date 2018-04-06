package command

import (
	"github.com/arquillian/ike-prow-plugins/pkg/github"
	gogh "github.com/google/go-github/github"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
)

// CommentCmdHandler keeps list of CommentCmd implementations to be handled when an IssueCommentEvent occurs
type CommentCmdHandler struct {
	Client          *github.Client
	commentCommands []CommentCmd
}

// Register adds the given CommentCmd implementation to the list of commands to be handled when an IssueCommentEvent occurs
func (s *CommentCmdHandler) Register(commentCommand CommentCmd) {
	s.commentCommands = append(s.commentCommands, commentCommand)
}

// Handle triggers the process of evaluating and performing of all stored CommentCmd implementations for the given comment
func (s *CommentCmdHandler) Handle(log log.Logger, prComment *gogh.IssueCommentEvent) error {
	for _, commentCommand := range s.commentCommands {
		if commentCommand.Matches(prComment) {
			err := commentCommand.Perform(s.Client, log, prComment)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// CommentCmd is a abstraction of a command that is triggered by a comment
type CommentCmd interface {
	// Perform triggers the process of evaluating and performing of the command for the given comment
	Perform(client *github.Client, log log.Logger, prComment *gogh.IssueCommentEvent) error
	// Matches says if the content of the given comment matches the command
	Matches(prComment *gogh.IssueCommentEvent) bool
}

package command

import (
	ghclient "github.com/arquillian/ike-prow-plugins/pkg/github/client"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	gogh "github.com/google/go-github/v41/github"
)

// CommentCmdHandler keeps list of CommentCmd implementations to be handled when an IssueCommentEvent occurs
type CommentCmdHandler struct {
	Client   ghclient.Client
	commands []CommentCmd
}

// Register adds the given CommentCmd implementation to the list of commands to be handled when an IssueCommentEvent occurs
func (s *CommentCmdHandler) Register(command CommentCmd) {
	s.commands = append(s.commands, command)
}

// Handle triggers the process of evaluating and performing of all stored CommentCmd implementations for the given comment
func (s *CommentCmdHandler) Handle(logger log.Logger, comment *gogh.IssueCommentEvent) error {
	for _, commentCommand := range s.commands {
		if commentCommand.Matches(comment) {
			err := commentCommand.Perform(s.Client, logger, comment)
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
	Perform(client ghclient.Client, logger log.Logger, comment *gogh.IssueCommentEvent) error
	// Matches says if the content of the given comment matches the command
	Matches(comment *gogh.IssueCommentEvent) bool
}

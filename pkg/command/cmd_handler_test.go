package command_test

import (
	"errors"

	is "github.com/arquillian/ike-prow-plugins/pkg/command"
	ghclient "github.com/arquillian/ike-prow-plugins/pkg/github/client"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	gogh "github.com/google/go-github/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type configurableCommentCommand struct {
	shouldMatch       bool
	shouldReturnError bool
	triggered         *bool
}

func (c *configurableCommentCommand) Perform(client ghclient.Client, logger log.Logger, comment *gogh.IssueCommentEvent) error {
	*c.triggered = true
	if c.shouldReturnError {
		return errors.New("error")
	}
	return nil
}

func (c *configurableCommentCommand) Matches(comment *gogh.IssueCommentEvent) bool {
	return c.shouldMatch
}

var _ = Describe("Command handler features", func() {

	client := NewDefaultGitHubClient()
	log := log.NewTestLogger()

	commentEvent := &gogh.IssueCommentEvent{
		Comment: &gogh.IssueComment{
			Body: utils.String("/command"),
		},
	}

	Context("Handling of registered commands", func() {

		It("should execute both commands when command name is matching in both cases", func() {
			// given
			firstTriggered := false
			secondTriggered := false
			firstCommand := &configurableCommentCommand{shouldMatch: true, triggered: &firstTriggered}
			secondCommand := &configurableCommentCommand{shouldMatch: true, triggered: &secondTriggered}

			commandHandler := is.CommentCmdHandler{Client: client}
			commandHandler.Register(firstCommand)
			commandHandler.Register(secondCommand)

			// when
			err := commandHandler.Handle(log, commentEvent)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(firstTriggered).To(BeTrue())
			Expect(secondTriggered).To(BeTrue())
		})

		It("should execute only first commands and return error thrown by the command", func() {
			// given
			firstTriggered := false
			secondTriggered := false
			firstCommand := &configurableCommentCommand{shouldMatch: true, shouldReturnError: true, triggered: &firstTriggered}
			secondCommand := &configurableCommentCommand{shouldMatch: true, triggered: &secondTriggered}

			commandHandler := is.CommentCmdHandler{Client: client}
			commandHandler.Register(firstCommand)
			commandHandler.Register(secondCommand)

			// when
			err := commandHandler.Handle(log, commentEvent)

			// then
			Ω(err).Should(HaveOccurred())
			Expect(firstTriggered).To(BeTrue())
			Expect(secondTriggered).To(BeFalse())
		})

		It("should execute only first commands because the second is not matching", func() {
			// given
			firstTriggered := false
			secondTriggered := false
			firstCommand := &configurableCommentCommand{shouldMatch: true, triggered: &firstTriggered}
			secondCommand := &configurableCommentCommand{shouldMatch: false, triggered: &secondTriggered}

			commandHandler := is.CommentCmdHandler{Client: client}
			commandHandler.Register(firstCommand)
			commandHandler.Register(secondCommand)

			// when
			err := commandHandler.Handle(log, commentEvent)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(firstTriggered).To(BeTrue())
			Expect(secondTriggered).To(BeFalse())
		})
	})
})

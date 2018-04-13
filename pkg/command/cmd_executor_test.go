package command_test

import (
	is "github.com/arquillian/ike-prow-plugins/pkg/command"
	"github.com/arquillian/ike-prow-plugins/pkg/github"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	gogh "github.com/google/go-github/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("Command executor features", func() {

	var (
		deletedCommand, triggeredCommand *gogh.IssueCommentEvent
		client                           = NewDefaultGitHubClient()
		log                              = NewDiscardOutLogger()
	)

	BeforeEach(func() {
		deletedCommand = &gogh.IssueCommentEvent{
			Comment: &gogh.IssueComment{
				Body: utils.String("/command"),
			},
			Action: utils.String("deleted"),
		}

		triggeredCommand = &gogh.IssueCommentEvent{
			Comment: &gogh.IssueComment{
				Body: utils.String("/command"),
			},
			Action: utils.String("created"),
			Issue: &gogh.Issue{
				Number: utils.Int(1),
			},
			Repo: &gogh.Repository{
				Name:  utils.String("repo"),
				Owner: &gogh.User{Login: utils.String("owner")},
			},
		}
	})

	Context("Executing of predefined commands in quiet mode", func() {

		It("should not execute command when command name is not matching", func() {
			// given
			executed := false
			command := is.CmdExecutor{Command: "/something-different", Quiet: true}
			command.When(is.Deleted).By(is.Anybody).Then(func() error {
				executed = true
				return nil
			})

			// when
			err := command.Execute(client, log, deletedCommand)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(executed).To(BeFalse())
		})

		It("should execute command when command name is matching anybody is allowed and deletion actions matches as well", func() {
			// given
			executed := false
			counter := 0
			command := is.CmdExecutor{Command: "/command", Quiet: true}
			command.When(is.Deleted).By(is.Anybody).Then(func() error {
				counter++
				executed = true
				return nil
			})

			// when
			err := command.Execute(client, log, deletedCommand)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(executed).To(BeTrue())
			Expect(counter).To(Equal(1))
		})

		It("should execute command once when multiple action are set and one is matching", func() {
			// given
			executed := false
			counter := 0
			command := is.CmdExecutor{Command: "/command", Quiet: true}
			command.When(is.Deleted, is.Triggered).By(is.Anybody).Then(func() error {
				counter++
				executed = true
				return nil
			})

			// when
			err := command.Execute(client, log, deletedCommand)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(executed).To(BeTrue())
			Expect(counter).To(Equal(1))
		})

		It("should execute command twice when multiple action are set and both are matching for two events", func() {
			// given
			executed := false
			counter := 0
			command := is.CmdExecutor{Command: "/command", Quiet: true}
			command.When(is.Deleted, is.Triggered).By(is.Anybody).Then(func() error {
				counter++
				executed = true
				return nil
			})

			// when
			err := command.Execute(client, log, deletedCommand)
			Ω(err).ShouldNot(HaveOccurred())
			err = command.Execute(client, log, triggeredCommand)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(executed).To(BeTrue())
			Expect(counter).To(Equal(2))
		})

		It("should not execute command when action is not matching", func() {
			// given
			executed := false
			command := is.CmdExecutor{Command: "/command", Quiet: true}
			command.When(is.Triggered).By(is.Anybody).Then(func() error {
				executed = true
				return nil
			})

			// when
			err := command.Execute(client, log, deletedCommand)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(executed).To(BeFalse())
		})

		It("should not execute command when restriction is not matching", func() {
			// given
			executed := false
			command := is.CmdExecutor{Command: "/command", Quiet: true}
			command.When(is.Deleted).By(is.Not(is.Anybody)).Then(func() error {
				executed = true
				return nil
			})

			// when
			err := command.Execute(client, log, deletedCommand)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(executed).To(BeFalse())
		})

		Context("Executing of predefined commands when comments are activated", func() {

			BeforeEach(func() {
				gock.OffAll()
			})

			It("should not execute command when command name is not matching", func() {
				// given

				triggeredCommand.Sender = &gogh.User{Login: utils.String("sender")}

				user := is.PermissionService{
					Client:   client,
					User:     "sender",
					PRLoader: github.NewPullRequestLazyLoader(client, triggeredCommand),
				}

				executed := false
				command := is.CmdExecutor{Command: "/command"}
				command.When(is.Triggered).By(user.Admin).Then(func() error {
					executed = true
					return nil
				})

				gock.New("https://api.github.com").
					Get("/repos/owner/repo/collaborators/sender/permission").
					Reply(200).
					BodyString("{\"permission\": \"read\"}")

				gock.New("https://api.github.com").
					Post("/repos/owner/repo/issues/1/comments").
					SetMatcher(
						ExpectPayload(To(
								HaveBodyThatContains("Hey @sender! It seems you tried to trigger `/command` command"),
								HaveBodyThatContains("You have to be admin")))).
					Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

				// when
				err := command.Execute(client, log, triggeredCommand)

				// then
				Ω(err).ShouldNot(HaveOccurred())
				Expect(executed).To(BeFalse())
			})

			It("should execute command because all conditions are fulfilled so no gock-not-matching-error is expected ", func() {
				// given
				executed := false
				command := is.CmdExecutor{Command: "/command"}
				command.When(is.Deleted).By(is.Anybody).Then(func() error {
					executed = true
					return nil
				})

				// when
				err := command.Execute(client, log, deletedCommand)

				// then
				Ω(err).ShouldNot(HaveOccurred())
				Expect(executed).To(BeTrue())
			})

			It("should not return gock-not-matching-error because the action is different so no request should be sent ", func() {
				// given
				executed := false
				command := is.CmdExecutor{Command: "/command"}
				command.When(is.Triggered).By(is.Anybody).Then(func() error {
					executed = true
					return nil
				})

				// when
				err := command.Execute(client, log, deletedCommand)

				// then
				Ω(err).ShouldNot(HaveOccurred())
				Expect(executed).To(BeFalse())
			})

			It("should return gock-not-matching-error when not quite mode and user doesn't have permissions", func() {
				// given
				executed := false
				command := is.CmdExecutor{Command: "/command"}
				command.When(is.Triggered).By(is.Not(is.Anybody)).Then(func() error {
					executed = true
					return nil
				})

				// when
				err := command.Execute(client, log, triggeredCommand)

				// then
				Ω(err).Should(HaveOccurred())
				Expect(executed).To(BeFalse())
			})
		})
	})
})

package github_test

import (
	"github.com/arquillian/ike-prow-plugins/pkg/github"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	gogh "github.com/google/go-github/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("GitHub Status Service", func() {

	Context("Updating PR statuses", func() {

		var statusService scm.StatusService

		toBe := func(status, description, context, targetURL string) func(statusPayload map[string]interface{}) bool {
			return func(statusPayload map[string]interface{}) bool {
				return Expect(statusPayload).To(SatisfyAll(
					HaveState(status),
					HaveDescription(description),
					HaveContext(context),
					HaveTargetURL(targetURL),
				))
			}
		}

		BeforeEach(func() {
			defer gock.Off()
			client := gogh.NewClient(nil)
			change := scm.RepositoryChange{RepoName: "test-repo", Owner: "alien-ike", Hash: "1232asdasd"}
			context := github.StatusContext{BotName: "alien-ike", PluginName: "test-keeper"}
			statusService = github.NewStatusService(client, NewDiscardOutLogger(), change, context)
		})

		It("should report success with context having bot name and plugin name", func() {
			// given
			gock.New("https://api.github.com").
				Post("/repos/alien-ike/test-repo/statuses/1232asdasd").
				SetMatcher(ExpectPayload(
						toBe(github.StatusSuccess, "All good, we have tests", "alien-ike/test-keeper", plugin.DocumentationURL))).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			// when
			err := statusService.Success("All good, we have tests", plugin.DocumentationURL)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should report failure with context and description", func() {
			// given
			gock.New("https://api.github.com").
				Post("/repos/alien-ike/test-repo/statuses/1232asdasd").
				SetMatcher(ExpectPayload(
						toBe(github.StatusFailure, "We don't have tests", "alien-ike/test-keeper", plugin.DocumentationURL))).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			// when
			err := statusService.Failure("We don't have tests", plugin.DocumentationURL)

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("should report pending without description", func() {
			// given
			gock.New("https://api.github.com").
				Post("/repos/alien-ike/test-repo/statuses/1232asdasd").
				SetMatcher(ExpectPayload(
						toBe(github.StatusPending, "", "alien-ike/test-keeper", ""))).
				Reply(201) // This way we implicitly verify that call happened after `HandleEvent` call

			// when
			err := statusService.Pending("")

			// then - implicit verification of /statuses call occurrence with proper payload
			Ω(err).ShouldNot(HaveOccurred())
		})

	})

})

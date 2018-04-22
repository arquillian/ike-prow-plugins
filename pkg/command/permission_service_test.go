package command_test

import (
	is "github.com/arquillian/ike-prow-plugins/pkg/command"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

var _ = Describe("Permission service with permission checks features", func() {

	Context("Getting and validating user's permissions", func() {

		client := NewDefaultGitHubClient()
		var user is.PermissionService

		BeforeEach(func() {
			gock.Off()
			user = is.PermissionService{
				Client: client,
				User:   "user",
				PRLoader: &ghservice.PullRequestLazyLoader{
					Client:    client,
					RepoOwner: "owner",
					RepoName:  "repo",
					Number:    1,
				},
			}
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should not approve the user when the permission is read", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/collaborators/user/permission").
				Reply(200).
				BodyString(`{"permission": "read"}`)

			// when
			status, err := user.Admin()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveRejectedUser("user"),
				HaveApprovedRoles(is.Admin),
				HaveNoRejectedRoles())
		})

		It("should approve the user when the permission is admin", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/collaborators/user/permission").
				Reply(200).
				BodyString(`{"permission": "admin"}`)

			// when
			status, err := user.Admin()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveApprovedUser("user"),
				HaveApprovedRoles(is.Admin),
				HaveNoRejectedRoles())
		})

		It("should not approve the user that is not the PR creator", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/pulls/1").
				Reply(200).
				BodyString(`{"user": {"login": "creator"}}`)

			// when
			status, err := user.PRCreator()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveRejectedUser("user"),
				HaveApprovedRoles(is.PullRequestCreator),
				HaveNoRejectedRoles())
		})

		It("should approve the user that is the PR creator", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/pulls/1").
				Reply(200).
				BodyString(`{"user": {"login": "user"}}`)

			// when
			status, err := user.PRCreator()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveApprovedUser("user"),
				HaveApprovedRoles(is.PullRequestCreator),
				HaveNoRejectedRoles())
		})

		It("should not approve the user that is not the requested PR reviewer", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/pulls/1").
				Reply(200).
				BodyString(`{"requested_reviewers": [{"login": "reviewer1"}, {"login": "reviewer2"}]}`)

			// when
			status, err := user.PRReviewer()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveRejectedUser("user"),
				HaveApprovedRoles(is.RequestReviewer),
				HaveNoRejectedRoles())
		})

		It("should approve the user that is one of the requested PR reviewers", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/pulls/1").
				Reply(200).
				BodyString(`{"requested_reviewers": [{"login": "reviewer1"}, {"login": "user"}]}`)

			// when
			status, err := user.PRReviewer()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveApprovedUser("user"),
				HaveApprovedRoles(is.RequestReviewer),
				HaveNoRejectedRoles())
		})

		It("should not approve the user that is not the PR approver", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/pulls/1/reviews").
				Reply(200).
				BodyString(`[{"user": {"login": "user"}, "state": "CHANGES_REQUESTED"},` +
					`{"user": {"login": "user"}, "state": "COMMENTED"}]`)

			// when
			status, err := user.PRApprover()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveRejectedUser("user"),
				HaveApprovedRoles(is.PullRequestApprover),
				HaveNoRejectedRoles())
		})

		It("should approve the user that is a PR approver", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/pulls/1/reviews").
				Reply(200).
				BodyString(`[{"user": {"login": "user"}, "state": "APPROVED"}]`)

			// when
			status, err := user.PRApprover()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveApprovedUser("user"),
				HaveApprovedRoles(is.PullRequestApprover),
				HaveNoRejectedRoles())
		})
	})

	Context("Permission check functions", func() {

		rejected := *is.NewPermissionStatus("user", false, []string{"role in rejected"}, []string{})
		approved := *is.NewPermissionStatus("user", true, []string{"role in approved"}, []string{})

		It("should approve everyone", func() {
			// when
			status, err := is.Anybody()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveApprovedUser(""),
				HaveApprovedRoles(is.Anyone),
				HaveNoRejectedRoles())
		})

		It("should approve everyone when no restrictions are set", func() {
			status, err := is.AnyOf()()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveApprovedUser(is.Unknown),
				HaveApprovedRoles(is.Anyone),
				HaveNoRejectedRoles())
		})

		It("should approve when at least one restriction is fulfilled", func() {
			// when
			status, err := is.AnyOf(newCheck(rejected), newCheck(approved))()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveApprovedUser("user"),
				HaveApprovedRoles("role in rejected", "role in approved"),
				HaveNoRejectedRoles())
		})

		It("should not approve when no restriction is fulfilled", func() {
			// given
			secondRejected := *is.NewPermissionStatus("user", false, []string{is.Admin}, []string{})

			// when
			status, err := is.AnyOf(newCheck(rejected), newCheck(secondRejected))()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveRejectedUser("user"),
				HaveApprovedRoles("role in rejected", is.Admin),
				HaveNoRejectedRoles())
		})

		It("should not approve when at least one restriction is not fulfilled", func() {
			// given
			secondApproved := *is.NewPermissionStatus("user", true, []string{is.Admin}, []string{})

			// when
			status, err := is.AllOf(newCheck(approved), newCheck(rejected), newCheck(secondApproved))()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveRejectedUser("user"),
				HaveApprovedRoles("role in approved", "role in rejected", is.Admin),
				HaveNoRejectedRoles())
		})

		It("should approve when all restrictions are fulfilled", func() {
			// given
			secondApproved := *is.NewPermissionStatus("user", true, []string{is.Admin}, []string{})

			// when
			status, err := is.AllOf(newCheck(approved), newCheck(secondApproved))()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveApprovedUser("user"),
				HaveApprovedRoles("role in approved", is.Admin),
				HaveNoRejectedRoles())
		})

		It("should reverse the approval to rejection", func() {
			// when
			status, err := is.Not(newCheck(approved))()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveRejectedUser("user"),
				HaveRejectedRoles("role in approved"),
				HaveNoApprovedRoles())
		})
	})
})

func newCheck(status is.PermissionStatus) func() (*is.PermissionStatus, error) {
	return func() (*is.PermissionStatus, error) {
		return &status, nil
	}
}

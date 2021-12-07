package command_test

import (
	is "github.com/arquillian/ike-prow-plugins/pkg/command"
	ghservice "github.com/arquillian/ike-prow-plugins/pkg/github/service"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gock "gopkg.in/h2non/gock.v1"
)

var _ = Describe("Permission service with permission checks features", func() {

	Context("Getting and validating user's permissions", func() {

		BeforeEach(func() {
			gock.Off()
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should not approve the user when the permission is read", func() {
			// given
			mock := MockPr().LoadedFromDefaultStruct().
				WithUsers(ExternalUser("user")).
				Create()

			// when
			status, err := mock.PermissionForUser("user").ThatIs().Admin(true)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveRejectedUser("user"),
				HaveApprovedRoles(is.Admin),
				HaveNoRejectedRoles())
		})

		It("should approve the user when the permission is admin", func() {
			// given
			mock := MockPr().LoadedFromDefaultStruct().
				WithUsers(Admin("user")).
				Create()

			// when
			status, err := mock.PermissionForUser("user").ThatIs().Admin(true)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveApprovedUser("user"),
				HaveApprovedRoles(is.Admin),
				HaveNoRejectedRoles())
		})

		It("should not approve the user that is not the PR creator", func() {
			// given
			mock := MockPr().LoadedFromDefaultStruct().
				WithUsers(PrCreator("creator")).
				Create()

			// when
			status, err := mock.PermissionForUser("user").ThatIs().PRCreator(true)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveRejectedUser("user"),
				HaveApprovedRoles(is.PullRequestCreator),
				HaveNoRejectedRoles())
		})

		It("should approve the user that is the PR creator", func() {
			// given
			mock := MockPr().LoadedFromDefaultStruct().
				WithUsers(PrCreator("user")).
				Create()

			// when
			status, err := mock.PermissionForUser("user").ThatIs().PRCreator(true)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveApprovedUser("user"),
				HaveApprovedRoles(is.PullRequestCreator),
				HaveNoRejectedRoles())
		})

		It("should not approve the user that is not the requested PR reviewer", func() {
			// given
			mock := MockPr().LoadedFromDefaultStruct().
				WithUsers(RequestedReviewer("reviewer1"), RequestedReviewer("reviewer2")).
				Create()

			// when
			status, err := mock.PermissionForUser("user").ThatIs().PRReviewer(true)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveRejectedUser("user"),
				HaveApprovedRoles(is.RequestedReviewer),
				HaveNoRejectedRoles())
		})

		It("should approve the user that is one of the requested PR reviewers", func() {
			// given
			mock := MockPr().LoadedFromDefaultStruct().
				WithUsers(RequestedReviewer("reviewer1"), RequestedReviewer("user")).
				Create()

			// when
			status, err := mock.PermissionForUser("user").ThatIs().PRReviewer(true)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveApprovedUser("user"),
				HaveApprovedRoles(is.RequestedReviewer),
				HaveNoRejectedRoles())
		})

		It("should not approve the user that is not the PR approver", func() {
			// given
			mock := MockPr().LoadedFromDefaultStruct().
				WithReviews(`[{"user": {"login": "user"}, "state": "CHANGES_REQUESTED"},` +
					`{"user": {"login": "user"}, "state": "COMMENTED"}]`).
				Create()

			// when
			status, err := mock.PermissionForUser("user").ThatIs().PRApprover(true)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveRejectedUser("user"),
				HaveApprovedRoles(is.PullRequestApprover),
				HaveNoRejectedRoles())
		})

		It("should approve the user that is a PR approver", func() {
			// given
			mock := MockPr().LoadedFromDefaultStruct().
				WithReviews(`[{"user": {"login": "user"}, "state": "APPROVED"}]`).
				Create()

			// when
			status, err := mock.PermissionForUser("user").ThatIs().PRApprover(true)

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
			status, err := is.Anybody(true)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveApprovedUser(""),
				HaveApprovedRoles(is.Anyone),
				HaveNoRejectedRoles())
		})

		It("should approve everyone when no restrictions are set", func() {
			status, err := is.AnyOf()(true)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveApprovedUser(is.Unknown),
				HaveApprovedRoles(is.Anyone),
				HaveNoRejectedRoles())
		})

		It("should approve when at least one restriction is fulfilled", func() {
			// when
			status, err := is.AnyOf(newCheck(rejected), newCheck(approved))(true)

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
			status, err := is.AnyOf(newCheck(rejected), newCheck(secondRejected))(true)

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
			status, err := is.AllOf(newCheck(approved), newCheck(rejected), newCheck(secondApproved))(true)

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
			status, err := is.AllOf(newCheck(approved), newCheck(secondApproved))(true)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveApprovedUser("user"),
				HaveApprovedRoles("role in approved", is.Admin),
				HaveNoRejectedRoles())
		})

		It("should reverse the approval to rejection", func() {
			// when
			status, err := is.Not(newCheck(approved))(true)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveRejectedUser("user"),
				HaveRejectedRoles("role in approved"),
				HaveNoApprovedRoles())
		})
	})

	Context("Lazy evaluation of the anyOff permission check function", func() {

		BeforeEach(func() {
			gock.Off()
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should approve based on first condition and not evaluate the remaining ones", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/collaborators/user/permission").
				Reply(200).
				BodyString(`{"permission": "read"}`)
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/pulls/1").
				Times(0).
				Reply(200)
			user := user()

			// when
			status, err := is.AnyOf(is.Not(user.Admin), user.PRReviewer, is.AllOf(user.PRApprover, user.PRCreator))(true)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveApprovedUser("user"),
				HaveApprovedRoles(is.RequestedReviewer, is.PullRequestApprover, is.PullRequestCreator),
				HaveRejectedRoles(is.Admin))
		})

		It("should approve based on first condition in nested anyOf check and not evaluate the remaining ones", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/pulls/1").
				Times(1).
				Reply(200).
				BodyString(`{"user": {"login": "user"}}`)
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/collaborators/user/permission").
				Times(0).
				Reply(200)
			user := user()

			// when
			status, err := is.AnyOf(user.PRReviewer, is.AnyOf(user.PRCreator, user.Admin), user.PRApprover)(true)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			ExpectPermissionStatus(status).To(
				HaveApprovedUser("user"),
				HaveApprovedRoles(is.Admin, is.RequestedReviewer, is.PullRequestApprover, is.PullRequestCreator),
				HaveNoRejectedRoles())
		})

	})
})

func user() *is.PermissionService {
	client := NewDefaultGitHubClient()
	return is.NewPermissionService(client, "user", &ghservice.PullRequestLazyLoader{
		Client:    client,
		RepoOwner: "owner",
		RepoName:  "repo",
		Number:    1,
	})
}

func newCheck(status is.PermissionStatus) func(evaluate bool) (*is.PermissionStatus, error) {
	return func(evaluate bool) (*is.PermissionStatus, error) {
		return &status, nil
	}
}

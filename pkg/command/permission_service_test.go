package command_test

import (
	is "github.com/arquillian/ike-prow-plugins/pkg/command"
	"github.com/arquillian/ike-prow-plugins/pkg/github"
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
				PRLoader: &github.PullRequestLoader{
					Client:    client,
					RepoOwner: "owner",
					RepoName:  "repo",
					Number:    1,
				},
			}
		})

		It("tests 'that is admin' function that should not approve the user when the permission is read", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/collaborators/user/permission").
				Reply(200).
				BodyString("{\"permission\": \"read\"}")

			// when
			permissionStatus, err := user.ThatIsAdmin()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			verifyStatusWithPredefinedUser(permissionStatus, false, "admin")
		})

		It("tests 'that is admin' function that should approve the user when the permission is admin", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/collaborators/user/permission").
				Reply(200).
				BodyString("{\"permission\": \"admin\"}")

			// when
			permissionStatus, err := user.ThatIsAdmin()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			verifyStatusWithPredefinedUser(permissionStatus, true, "admin")
		})

		It("tests 'that is pr creator' function that should not approve the user that is not the PR creator", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/pulls/1").
				Reply(200).
				BodyString(`{"user": {"login": "creator"}}`)

			// when
			permissionStatus, err := user.ThatIsPRCreator()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			verifyStatusWithPredefinedUser(permissionStatus, false, "pull request creator")
		})

		It("tests 'that is pr creator' function that should approve the user that is the PR creator", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/pulls/1").
				Reply(200).
				BodyString(`{"user": {"login": "user"}}`)

			// when
			permissionStatus, err := user.ThatIsPRCreator()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			verifyStatusWithPredefinedUser(permissionStatus, true, "pull request creator")
		})

		It("tests 'that is pr reviewer' function that should not approve the user that is not the requested PR reviewer", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/pulls/1").
				Reply(200).
				BodyString(`{"requested_reviewers": [{"login": "reviewer1"}, {"login": "reviewer2"}]}`)

			// when
			permissionStatus, err := user.ThatIsPRReviewer()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			verifyStatusWithPredefinedUser(permissionStatus, false, "requested reviewer")
		})

		It("tests 'that is pr reviewer' function that should approve the user that is one of the requested PR reviewers", func() {
			// given
			gock.New("https://api.github.com").
				Get("/repos/owner/repo/pulls/1").
				Reply(200).
				BodyString(`{"requested_reviewers": [{"login": "reviewer1"}, {"login": "user"}]}`)

			// when
			permissionStatus, err := user.ThatIsPRReviewer()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			verifyStatusWithPredefinedUser(permissionStatus, true, "requested reviewer")
		})
	})

	Context("Permission check functions", func() {

		rejected := is.PermissionStatus{
			User:           "user",
			UserIsApproved: false,
			ApprovedRoles:  []string{"role in rejected"},
		}
		approved := is.PermissionStatus{
			User:           "user",
			UserIsApproved: true,
			ApprovedRoles:  []string{"role in approved"},
		}

		It("tests 'anybody' function that should approve everyone", func() {
			// when
			permissionStatus, err := is.AnyBody()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			verifyStatus(permissionStatus, "", true, "anyone")
		})

		It("tests 'anyOf' function that should approve everyone when no restrictions are set", func() {
			permissionStatus, err := is.AnyOf()()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			verifyStatus(permissionStatus, "unknown", true, "anyone")
		})

		It("tests 'AnyOf' function that should approve when at least one restriction is fulfilled", func() {
			// given

			// when
			permissionStatus, err := is.AnyOf(newCheck(rejected), newCheck(approved))()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			verifyStatus(permissionStatus, "user", true, "role in rejected", "role in approved")
		})

		It("tests 'AnyOf' function that should not approve when no restriction is fulfilled", func() {
			// when
			permissionStatus, err := is.AnyOf(newCheck(rejected), newCheck(rejected))()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			verifyStatus(permissionStatus, "user", false, "role in rejected", "role in rejected")
		})

		It("tests 'AllOf' function that should not approve when at least one restriction is not fulfilled", func() {
			// when
			permissionStatus, err := is.AllOf(newCheck(approved), newCheck(rejected), newCheck(approved))()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			verifyStatus(permissionStatus, "user", false, "role in approved", "role in rejected", "role in approved")
		})

		It("tests 'AllOf' function that should approve when all restrictions are fulfilled", func() {
			// given

			// when
			permissionStatus, err := is.AllOf(newCheck(approved), newCheck(approved))()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			verifyStatus(permissionStatus, "user", true, "role in approved", "role in approved")
		})

		It("tests 'Not' function that should reverse the approval to rejection", func() {
			// when
			status, err := is.Not(newCheck(approved))()

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(status.User).To(Equal("user"))
			Expect(status.UserIsApproved).To(Equal(false))
			Expect(status.ApprovedRoles).To(BeEmpty())
			Expect(status.RejectedRoles).To(ConsistOf("role in approved"))
		})
	})
})

func verifyStatusWithPredefinedUser(status *is.PermissionStatus, userIsApproved bool, approvedRole ...string) {
	verifyStatus(status, "user", userIsApproved, approvedRole...)
}

func verifyStatus(status *is.PermissionStatus, userName string, userIsApproved bool, approvedRole ...string) {
	Expect(status.User).To(Equal(userName))
	Expect(status.UserIsApproved).To(Equal(userIsApproved))
	Expect(status.ApprovedRoles).To(ConsistOf(approvedRole))
	Expect(status.RejectedRoles).To(BeEmpty())
}

func newCheck(status is.PermissionStatus) func() (*is.PermissionStatus, error) {
	return func() (*is.PermissionStatus, error) {
		return &status, nil
	}
}

package ghservice_test

import (
	ghservice "github.com/arquillian/ike-prow-plugins/pkg/github/service"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	"github.com/google/go-github/v41/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gock "gopkg.in/h2non/gock.v1"
)

var _ = Describe("Issue comments lazy loading", func() {

	client := NewDefaultGitHubClient()

	BeforeEach(func() {
		defer gock.OffAll()
	})

	AfterEach(EnsureGockRequestsHaveBeenMatched)

	It("should load issue comments when load() method is called", func() {
		// given
		calls := 0
		gock.New("https://api.github.com").
			Get("/repos/owner/name/issues/123/comments").
			SetMatcher(SpyOnCalls(&calls)).
			Reply(200).
			BodyString(`[{"user":{"login":"commenter"}, "body":"cool comment"}]`)
		issue := scm.NewRepositoryIssue("owner", "name", 123)
		loader := &ghservice.IssueCommentsLazyLoader{Client: client, Issue: *issue}
		Expect(calls).To(Equal(0))

		// when
		comments, err := loader.Load()

		// then
		Ω(err).ShouldNot(HaveOccurred())
		Expect(calls).To(Equal(1))
		expComment := &github.IssueComment{
			User: &github.User{Login: utils.String("commenter")},
			Body: utils.String("cool comment"),
		}
		Expect(comments).To(ConsistOf(expComment))
	})

	It("should load issue comments only once", func() {
		// given
		counter := 0
		gock.New("https://api.github.com").
			Get("/repos/owner/name/issues/123/comments").
			SetMatcher(SpyOnCalls(&counter)).
			Persist().
			Reply(200).
			BodyString(`[{"user":{"login":"commenter"}, "body":"cool comment"}]`)
		issue := scm.NewRepositoryIssue("owner", "name", 123)
		loader := &ghservice.IssueCommentsLazyLoader{Client: client, Issue: *issue}
		_, _ = loader.Load()

		// when
		comments, err := loader.Load()

		// then
		Ω(err).ShouldNot(HaveOccurred())
		Expect(counter).To(Equal(1))
		expComment := &github.IssueComment{
			User: &github.User{Login: utils.String("commenter")},
			Body: utils.String("cool comment"),
		}
		Expect(comments).To(ConsistOf(expComment))
	})
})

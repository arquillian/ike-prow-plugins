package ghclient_test

import (
	ghclient "github.com/arquillian/ike-prow-plugins/pkg/github/client"
	. "github.com/arquillian/ike-prow-plugins/pkg/internal/test"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	gogh "github.com/google/go-github/v41/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gock "gopkg.in/h2non/gock.v1"
)

var _ = Describe("Pagination checker", func() {

	const repositoryName = "bartoszmajsak/wfswarm-booster-pipeline-test"
	client := ghclient.NewClient(gogh.NewClient(nil), log.NewTestLogger())
	client.RegisterAroundFunctions(
		ghclient.NewRateLimitWatcher(client, log.NewTestLogger(), 100),
		ghclient.NewRetryWrapper(3, 0),
		ghclient.NewPaginationChecker())

	Context("Pagination checker should correctly detect when there are some more pages available", func() {

		BeforeEach(func() {
			defer gock.OffAll()
		})

		AfterEach(EnsureGockRequestsHaveBeenMatched)

		It("should get all 3 pages and group the entries together", func() {
			// given
			mockHighRateLimit()
			gock.New("https://api.github.com").
				Get("/repos/"+repositoryName+"/pulls/2/files").
				MatchParam("per_page", "100").
				MatchParam("page", "1").
				Reply(200).
				Body(FromFile("test_fixtures/gh/list_files_page_1.json")).
				AddHeader("Link",
					"<https://api.github.com/repositories/121737972/pulls/2/files?per_page=1&page=2>; rel=\"next\", "+
						"<https://api.github.com/repositories/121737972/pulls/2/files?per_page=1&page=3>; rel=\"last\"")

			gock.New("https://api.github.com").
				Get("/repos/"+repositoryName+"/pulls/2/files").
				MatchParam("per_page", "100").
				MatchParam("page", "2").
				Reply(200).
				Body(FromFile("test_fixtures/gh/list_files_page_2.json")).
				AddHeader("Link",
					"<https://api.github.com/repositories/121737972/pulls/2/files?per_page=1&page=1>; rel=\"prev\", "+
						"<https://api.github.com/repositories/121737972/pulls/2/files?per_page=1&page=3>; rel=\"next\", "+
						"<https://api.github.com/repositories/121737972/pulls/2/files?per_page=1&page=3>; rel=\"last\", "+
						"<https://api.github.com/repositories/121737972/pulls/2/files?per_page=1&page=1>; rel=\"first\"")

			gock.New("https://api.github.com").
				Get("/repos/"+repositoryName+"/pulls/2/files").
				MatchParam("per_page", "100").
				MatchParam("page", "3").
				Reply(200).
				Body(FromFile("test_fixtures/gh/list_files_page_3.json")).
				AddHeader("Link",
					"<https://api.github.com/repositories/121737972/pulls/2/files?per_page=1&page=2>; rel=\"prev\", "+
						"<https://api.github.com/repositories/121737972/pulls/2/files?per_page=1&page=1>; rel=\"first\"")

			// when
			files, err := client.ListPullRequestFiles("bartoszmajsak", "wfswarm-booster-pipeline-test", 2)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(gock.GetUnmatchedRequests()).To(BeEmpty())
			Expect(files).To(HaveLen(3))
			Expect(files).To(ConsistOf(
				newChangedFile("Jenkinsfile", "modified", 3, 3),
				newChangedFile("README.adoc", "modified", 2, 2),
				newChangedFile("src/test/java/io/openshift/booster/NewTest.java", "added", 66, 0),
			))

		})
	})
})

func newChangedFile(name, status string, additions, deletions int) scm.ChangedFile {
	return scm.ChangedFile{
		Name:      name,
		Status:    status,
		Additions: additions,
		Deletions: deletions,
	}
}

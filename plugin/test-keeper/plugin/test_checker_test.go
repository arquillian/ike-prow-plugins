package plugin_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/arquillian/ike-prow-plugins/scm"
	"github.com/arquillian/ike-prow-plugins/plugin/test-keeper/plugin"
)

type StubCommitScmService struct {
	CommitScmService
	repoService    RepoScmService
	affectedFiles  []AffectedFile
	rawFileContent string
}

type StubRepoScmService struct {
	RepoScmService
	languages []string
}

func (stub StubCommitScmService) GetRepoService() RepoScmService {
	return stub.repoService
}

func (stub StubCommitScmService) GetAffectedFiles() ([]AffectedFile, error) {
	return stub.affectedFiles, nil
}

func (stub StubCommitScmService) GetRawFile(filePath string) []byte {
	return []byte(stub.rawFileContent)
}

func (stub StubRepoScmService) GetRepoLanguages() ([]string, error) {
	return stub.languages, nil
}

func createStubCommitScmService(languages []string, rawFileContent string, affectedFiles []AffectedFile) StubCommitScmService {
	emptyCommitService := &CommitScmServiceImpl{}
	return StubCommitScmService{
		CommitScmService: emptyCommitService,
		repoService: &StubRepoScmService{
			RepoScmService: emptyCommitService.GetRepoService(),
			languages:      languages,
		},
		affectedFiles:  affectedFiles,
		rawFileContent: rawFileContent,
	}
}

var _ = Describe("Test Checker features", func() {

	Context("Test Checker lookup for the test files within a commit", func() {

		It("should return true when a Java test file is affected within a commit", func() {
			// given
			stub := createStubCommitScmService([]string{"Java", "JavaScript", "HTML"}, "",
				createAddedAffectedFiles(
					"path/to/Anything.java",
					"path/to/page.html",
					"path/to/test/AnythingTestCase.java"))

			checker := plugin.TestChecker{Log: Logger, CommitService: stub}

			// when
			bool, err := checker.IsAnyTestPresent()

			// then
			Expect(err).Should(BeNil())
			Expect(bool).Should(BeTrue())
		})

		It("should return false when no test file is affected within a commit", func() {
			// given
			stub := createStubCommitScmService([]string{"Java", "JavaScript", "HTML"}, "",
				createAddedAffectedFiles(
					"path/to/Anything.java",
					"path/to/page.html",
					"path/to/js/something.in.js"))

			checker := plugin.TestChecker{Log: Logger, CommitService: stub}

			// when
			bool, err := checker.IsAnyTestPresent()

			// then
			Expect(err).Should(BeNil())
			Expect(bool).Should(BeFalse())
		})

		It("should return true when test file is matched using custom config", func() {
			// given
			stub := createStubCommitScmService([]string{"Go", "HTML"}, "tests_pattern: .*tezt.my",
				createAddedAffectedFiles(
					"path/to/page.html",
					"path/to/custom.tezt.my"))

			checker := plugin.TestChecker{Log: Logger, CommitService: stub}

			// when
			bool, err := checker.IsAnyTestPresent()

			// then
			Expect(err).Should(BeNil())
			Expect(bool).Should(BeTrue())
		})

		It("should return false when test file is not matched using custom config", func() {
			// given
			stub := createStubCommitScmService([]string{"Go", "Java"}, "tests_pattern: .*tezt.my",
				createAddedAffectedFiles(
					"path/to/MyTestCase.java",
					"path/to/another_test.go"))

			checker := plugin.TestChecker{Log: Logger, CommitService: stub}

			// when
			bool, err := checker.IsAnyTestPresent()

			// then
			Expect(err).Should(BeNil())
			Expect(bool).Should(BeFalse())
		})
	})

})

func createAddedAffectedFiles(names ...string) []AffectedFile {
	files := make([]AffectedFile, len(names))
	for _, name := range names {
		files = append(files, AffectedFile{Name: name, Status: "added"})
	}
	return files
}

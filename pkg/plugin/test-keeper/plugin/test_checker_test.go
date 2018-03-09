package plugin_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	. "github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper/plugin"
	"github.com/arquillian/ike-prow-plugins/pkg/internal/test"
)

var _ = Describe("Test Checker features", func() {

	Context("Test Checker lookup for the test files within a commit", func() {

		It("should find tests in the Java file set when matchers for Java are explicitly defined", func() {
			// given
			matchers := LoadMatchers(TestKeeperConfiguration{}, func() []string {
				return []string{"Go", "Java"}
			})

			changedFiles := createChangedFiles(
				"path/to/Anything.java",
				"path/to/page.html",
				"path/to/test/AnythingTestCase.java")

			checker := TestChecker{Log: test.CreateNullLogger(), TestMatchers: matchers}

			// when
			testsExist, err := checker.IsAnyTestPresent(changedFiles)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(testsExist).To(BeTrue())
		})

		It("should return false when no test file is affected within a commit", func() {
			// given
			matchers := LoadMatchers(TestKeeperConfiguration{}, func() []string {
				return []string{"Go", "Java"}
			})
			changedFiles := createChangedFiles(
					"path/to/Anything.java",
					"path/to/page.html",
					"path/to/js/something.in.js",
					"path/to/go/another_in.go")

			checker := TestChecker{Log: test.CreateNullLogger(), TestMatchers: matchers}

			// when
			testsExist, err := checker.IsAnyTestPresent(changedFiles)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(testsExist).To(BeFalse())
		})
	})

})

// nolint
func createChangedFiles(names ...string) []scm.ChangedFile {
	files := make([]scm.ChangedFile, len(names))
	for _, name := range names {
		files = append(files, scm.ChangedFile{Name: name, Status: "added"})
	}
	return files
}

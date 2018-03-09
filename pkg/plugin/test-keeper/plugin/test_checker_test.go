package plugin

import (
	. "github.com/onsi/ginkgo"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
)

var _ = Describe("Test Checker features", func() {

	Context("Test Checker lookup for the test files within a commit", func() {

		It("should return true when a Java test file is affected within a commit", func() {
			//// given
			//stub := createStubCommitScmService([]string{"Java", "JavaScript", "HTML"}, "",
			//	createAddedAffectedFiles(
			//		"path/to/Anything.java",
			//		"path/to/page.html",
			//		"path/to/test/AnythingTestCase.java"))
			//
			//checker := plugin.TestChecker{Log: test.CreateNullLogger(), CommitService: stub}
			//
			//// when
			//bool, err := checker.IsAnyTestPresent()
			//
			//// then
			//Expect(err).To(BeNil())
			//Expect(bool).To(BeTrue())
		})

		It("should return false when no test file is affected within a commit", func() {
			//// given
			//stub := createStubCommitScmService([]string{"Java", "JavaScript", "HTML"}, "",
			//	createAddedAffectedFiles(
			//		"path/to/Anything.java",
			//		"path/to/page.html",
			//		"path/to/js/something.in.js"))
			//
			//checker := plugin.TestChecker{Log: test.CreateNullLogger(), CommitService: stub}
			//
			//// when
			//bool, err := checker.IsAnyTestPresent()
			//
			//// then
			//Expect(err).To(BeNil())
			//Expect(bool).To(BeFalse())
		})

		It("should return true when test file is matched using custom config", func() {
			//// given
			//stub := createStubCommitScmService([]string{"Go", "HTML"}, "tests_pattern: .*tezt.my",
			//	createAddedAffectedFiles(
			//		"path/to/page.html",
			//		"path/to/custom.tezt.my"))
			//
			//checker := plugin.TestChecker{Log: test.CreateNullLogger(), CommitService: stub}
			//
			//// when
			//bool, err := checker.IsAnyTestPresent()
			//
			//// then
			//Expect(err).To(BeNil())
			//Expect(bool).To(BeTrue())
		})

		It("should return false when test file is not matched using custom config", func() {
			//// given
			//stub := createStubCommitScmService([]string{"Go", "Java"}, "tests_pattern: .*tezt.my",
			//	createAddedAffectedFiles(
			//		"path/to/MyTestCase.java",
			//		"path/to/another_test.go"))
			//
			//checker := plugin.TestChecker{Log: test.CreateNullLogger(), CommitService: stub}
			//
			//// when
			//bool, err := checker.IsAnyTestPresent()
			//
			//// then
			//Expect(err).To(BeNil())
			//Expect(bool).To(BeFalse())
		})
	})

})

func createAddedAffectedFiles(names ...string) []scm.ChangedFile {
	files := make([]scm.ChangedFile, len(names))
	for _, name := range names {
		files = append(files, scm.ChangedFile{Name: name, Status: "added"})
	}
	return files
}

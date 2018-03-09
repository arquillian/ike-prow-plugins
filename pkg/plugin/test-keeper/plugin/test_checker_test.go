package plugin_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	. "github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper/plugin"
	"github.com/arquillian/ike-prow-plugins/pkg/internal/test"
)

var _ = Describe("Test Checker features", func() {

	Context("Detecting tests within file changeset", func() {

		It("should find tests in the Java file set when matchers for Java are explicitly defined", func() {
			// given
			matchers := LoadMatchers(TestKeeperConfiguration{}, func() []string {
				return []string{"Go", "Java"}
			})

			changedFiles := changedFilesSet(
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

		It("should find Go and Java tests using predefined matchers based on languages in repository", func() {
			// given
			matchers := LoadMatchers(TestKeeperConfiguration{}, func() []string {
				return []string{"Go", "Java"}
			})

			changedFiles := changedFilesSet(
				"path/to/JavaTest.java",
				"path/to/golang/main_test.go")

			checker := TestChecker{Log: test.CreateNullLogger(), TestMatchers: matchers}

			// when
			testsExist, err := checker.IsAnyTestPresent(changedFiles)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(testsExist).To(BeTrue())
		})

		It("should find Java tests using predefined matchers based on languages in repository", func() {
			// given
			matchers := LoadMatchers(TestKeeperConfiguration{}, func() []string {
				return []string{"Java"}
			})

			changedFiles := changedFilesSet(
				"src/test/java/JavaTest.java",
				"path/to/golang/main_test.go")

			checker := TestChecker{Log: test.CreateNullLogger(), TestMatchers: matchers}

			// when
			testsExist, err := checker.IsAnyTestPresent(changedFiles)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(testsExist).To(BeTrue())
		})

		It("should find Go tests using predefined matchers based on languages in repository", func() {
			// given
			matchers := LoadMatchers(TestKeeperConfiguration{}, func() []string {
				return []string{"Go"}
			})

			changedFiles := changedFilesSet(
				"pkg/plugin/test-keeper/plugin/test_checker.go",
				"path/to/golang/main_test.go")

			checker := TestChecker{Log: test.CreateNullLogger(), TestMatchers: matchers}

			// when
			testsExist, err := checker.IsAnyTestPresent(changedFiles)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(testsExist).To(BeTrue())
		})

		// This test is failing ;)
		It("should not detect any tests when files are not matching predefined language patterns", func() {
			// given
			matchers := LoadMatchers(TestKeeperConfiguration{}, func() []string {
				return []string{"Go", "Java"}
			})
			changedFiles := changedFilesSet(
					"path/to/Anything.java",
					"path/_test.go/page.html",
					"path/Test.java/js/something.in.js",
					"path/to/go/another_in.go")

			checker := TestChecker{Log: test.CreateNullLogger(), TestMatchers: matchers}

			// when
			testsExist, err := checker.IsAnyTestPresent(changedFiles)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(testsExist).To(BeFalse())
		})

		It("should not detect any tests when change set is empty", func() {
			// given
			matchers := LoadMatchers(TestKeeperConfiguration{}, func() []string {
				return []string{"Go", "Java"}
			})
			changedFiles := changedFilesSet()

			checker := TestChecker{Log: test.CreateNullLogger(), TestMatchers: matchers}

			// when
			testsExist, err := checker.IsAnyTestPresent(changedFiles)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(testsExist).To(BeFalse())
		})

		It("should find tests using inclusion in the configuration", func() {
			// given
			matchers := LoadMatchers(TestKeeperConfiguration{
				Inclusion: `_test\.rb$`,
			}, func() []string {
				return []string{"Ruby"}
			})

			changedFiles := changedFilesSet(
				"path/to/github_service.rb",
				"path/to/github_service_test.rb")

			checker := TestChecker{Log: test.CreateNullLogger(), TestMatchers: matchers}

			// when
			testsExist, err := checker.IsAnyTestPresent(changedFiles)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(testsExist).To(BeTrue())
		})

		It("should find tests using inclusion in the configuration", func() {
			// given
			matchers := LoadMatchers(TestKeeperConfiguration{
				Inclusion: `(Test\.java|TestCase\.java|_test\.go)$`,
			}, func() []string {
				return []string{}
			})

			changedFiles := changedFilesSet(
				"path/to/JavaTest.java",
				"path/to/JavaTestCase.java",
				"path/to/JavaTestCase.java",
				"path/to/golang/main_test.go")

			checker := TestChecker{Log: test.CreateNullLogger(), TestMatchers: matchers}

			// when
			testsExist, err := checker.IsAnyTestPresent(changedFiles)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(testsExist).To(BeTrue())
		})
	})

})

func changedFilesSet(names ...string) []scm.ChangedFile {
	files := make([]scm.ChangedFile, len(names))
	for _, name := range names {
		files = append(files, scm.ChangedFile{Name: name, Status: "added"})
	}
	return files
}

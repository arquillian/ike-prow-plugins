package plugin_test

import (
	"regexp"

	"github.com/arquillian/ike-prow-plugins/pkg/plugin/test-keeper/plugin"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test keeper comment message parsing", func() {

	Context("Plugin hint file matcher", func() {

		It("should match plugin hint file from relative path", func() {
			// given
			pluginHint := "path/to/custom_message_file.md"

			// when
			isFilePath, _ := regexp.MatchString(plugin.FileRegex, pluginHint)

			// then
			Expect(isFilePath).To(Equal(true))
		})

		It("should match plugin hint file from url", func() {
			// given
			pluginHint := "http://my.server.com/path/to/custom_message_file.md"

			// when
			isFilePath, _ := regexp.MatchString(plugin.FileRegex, pluginHint)

			// then
			Expect(isFilePath).To(Equal(true))
		})

		It("should match plugin hint file from secure url", func() {
			// given
			pluginHint := "https://my.server.com/path/to/custom_message_file.md"

			// when
			isFilePath, _ := regexp.MatchString(plugin.FileRegex, pluginHint)

			// then
			Expect(isFilePath).To(Equal(true))
		})

		It("should match plugin hint file with upper case file extension", func() {
			// given
			pluginHint := "http://my.server.com/path/to/custom_message_file.MD"

			// when
			isFilePath, _ := regexp.MatchString(plugin.FileRegex, pluginHint)

			// then
			Expect(isFilePath).To(Equal(true))
		})

		It("should not match plugin hint inline comment", func() {
			// given
			pluginHint := "Custom message."

			// when
			isFilePath, _ := regexp.MatchString(plugin.FileRegex, pluginHint)

			// then
			Expect(isFilePath).To(Equal(false))
		})
	})
})

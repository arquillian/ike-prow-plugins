package config_test

import (
	"github.com/arquillian/ike-prow-plugins/pkg/plugin/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

type sampleConfiguration struct {
	config.PluginConfiguration `yaml:",inline,omitempty"`
	Name                       string `yaml:"name,omitempty"`
}

type testConfigProvider func() []config.Source

func (s testConfigProvider) Sources() []config.Source {
	return s()
}

var onlyName = config.Source(func() ([]byte, error) {
	return []byte("name: 'awesome-o'"), nil
})

var nameAndHint = config.Source(func() ([]byte, error) {
	return []byte("name: 'name-and-hint'\n" +
		"plugin_hint: 'I am just a hint'"), nil
})

var faulty = config.Source(func() ([]byte, error) {
	return nil, errors.New("no config found here")
})

var _ = Describe("Config loader features", func() {

	Context("Loading configuration using different strategies", func() {

		It("should load configuration when a successful lookup provided", func() {
			// given
			testConfigProviders := testConfigProvider(func() []config.Source {
				return []config.Source{onlyName}
			})

			sampleConfig := sampleConfiguration{}

			// when
			err := config.Load(&sampleConfig, testConfigProviders)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(sampleConfig.Name).To(Equal("awesome-o"))
		})

		It("should load configuration when failing and successful lookup provided, skipping first failing", func() {
			// given
			testConfigProviders := testConfigProvider(func() []config.Source {
				return []config.Source{faulty, onlyName}
			})

			sampleConfig := sampleConfiguration{}

			// when
			err := config.Load(&sampleConfig, testConfigProviders)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(sampleConfig.Name).To(Equal("awesome-o"))
		})

		It("should load config from first working source (precedence)", func() {
			// given
			testConfigProviders := testConfigProvider(func() []config.Source {
				return []config.Source{nameAndHint, onlyName}
			})

			sampleConfig := sampleConfiguration{Name: "prototype"}

			// when
			err := config.Load(&sampleConfig, testConfigProviders)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(sampleConfig.Name).To(Equal("name-and-hint"))
			Expect(sampleConfig.PluginHint).To(Equal("I am just a hint"))
		})

		It("should preserve prototype config name when no sources provided", func() {
			// given
			testConfigProviders := testConfigProvider(func() []config.Source {
				return []config.Source{}
			})

			sampleConfig := sampleConfiguration{Name: "prototype"}

			// when
			err := config.Load(&sampleConfig, testConfigProviders)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(sampleConfig.Name).To(Equal("prototype"))
		})

		It("should not propagate error when faulty source provided", func() {
			// given
			testConfigProviders := testConfigProvider(func() []config.Source {
				return []config.Source{faulty}
			})

			sampleConfig := sampleConfiguration{Name: "prototype"}

			// when
			err := config.Load(&sampleConfig, testConfigProviders)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(sampleConfig.Name).To(Equal("prototype"))
		})

	})
})

package http_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/arquillian/ike-prow-plugins/pkg/http"
	"time"
	gogh "github.com/google/go-github/github"
	"net/http"
	"errors"
	. "github.com/onsi/gomega"
)

var _ = Describe("http retry function", func() {

	Context("Retrying http requests", func() {

		It("should retry once when invoker returns error", func() {
			// given
			executionCounter := 0
			requestInvoker := func() (*gogh.Response, error) {
				executionCounter++
				return &gogh.Response{Response: &http.Response{StatusCode: 401}}, errors.New("unauthorized")
			}

			// when
			err := Do(4, time.Microsecond, requestInvoker)

			// then
			Ω(err).Should(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("All 1 attempts of sending a request failed"))
			Expect(executionCounter).To(Equal(1))
		})

		It("should retry 3 times when invoker returns 404 twice and 401 for the third attempt", func() {
			// given
			executionCounter := 0
			requestInvoker := func() (*gogh.Response, error) {
				executionCounter++
				if executionCounter == 3 {
					return &gogh.Response{Response: &http.Response{StatusCode: 401}}, errors.New("unauthorized")
				}
				return &gogh.Response{Response: &http.Response{StatusCode: 404}}, errors.New("not found")
			}

			// when
			err := Do(10, time.Microsecond, requestInvoker)

			// then
			Ω(err).Should(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("All 3 attempts of sending a request failed"))
			Expect(executionCounter).To(Equal(3))
		})

		It("should retry for all 4 times when invoker returns only 404 without any error", func() {
			// given
			executionCounter := 0
			requestInvoker := func() (*gogh.Response, error) {
				executionCounter++
				return &gogh.Response{Response: &http.Response{StatusCode: 404}}, nil
			}

			// when
			err := Do(4, time.Microsecond, requestInvoker)

			// then
			Ω(err).Should(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("All 4 attempts of sending a request failed"))
			Expect(err.Error()).To(ContainSubstring("server responded with error 404"))
			Expect(executionCounter).To(Equal(4))
		})

		It("should retry 3 times and not fail when invoker returns 404 twice and 200 for the third attempt", func() {
			// given
			executionCounter := 0
			requestInvoker := func() (*gogh.Response, error) {
				executionCounter++
				if executionCounter == 3 {
					return &gogh.Response{Response: &http.Response{StatusCode: 200}}, nil
				}
				return &gogh.Response{Response: &http.Response{StatusCode: 404}}, errors.New("not found")
			}

			// when
			err := Do(10, time.Microsecond, requestInvoker)

			// then
			Ω(err).ShouldNot(HaveOccurred())
			Expect(executionCounter).To(Equal(3))
		})
	})

})

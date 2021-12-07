package retry_test

import (
	"errors"

	"github.com/arquillian/ike-prow-plugins/pkg/retry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Retry logic", func() {

	It("should accumulate all errors when all attempts failed", func() {
		// given
		maxRetries := 4
		executions := 0
		toRetry := func() error {
			executions++
			return errors.New("unauthorized")
		}

		// when
		err := retry.Do(maxRetries, 0, toRetry)

		// then
		Expect(err).To(HaveLen(maxRetries))
		Expect(executions).To(Equal(maxRetries))
	})

	It("should execute once when 0 retries specified", func() {
		// given
		maxRetries := 0
		executions := 0
		toRetry := func() error {
			executions++
			return errors.New("unauthorized")
		}

		// when
		err := retry.Do(maxRetries, 0, toRetry)

		// then
		Expect(err).To(HaveLen(1))
		Expect(executions).To(Equal(1))
	})

	It("should retry 3 times out of 10 times when 3rd attempt successful and don't return errors", func() {
		// given
		executions := 0
		toRetry := func() error {
			executions++
			if executions == 3 {
				return nil
			}
			return errors.New("not found")
		}

		// when
		err := retry.Do(10, 0, toRetry)

		// then
		Expect(err).To(BeEmpty())
		Expect(executions).To(Equal(3))
	})

})

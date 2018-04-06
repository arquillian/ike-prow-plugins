package utils

import (
	"time"
)

// OnRetryFunction is a function that returns a boolean whether it should be invoked again or not and a possible error
type OnRetryFunction func() (shouldRetry bool, error error)

// Retry invokes a function for the given amount of retries with the given sleep before each one of them
func Retry(retries int, sleep time.Duration, onRetry OnRetryFunction) []error {
	errs := make([]error, 0, retries)

	shouldRetry, err := onRetry()

	for i := 0; i < retries-1 && shouldRetry; i++ {
		errs = append(errs, err)
		time.Sleep(sleep)
		shouldRetry, err = onRetry()
	}

	if err != nil {
		errs = append(errs, err)
		return errs
	}
	return make([]error, 0)
}

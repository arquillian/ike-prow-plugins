package retry

import (
	"time"
)

// Do invokes a function for the given amount of retries with the given sleep before each one of them
func Do(retries int, sleep time.Duration, onRetry func() error) []error {
	errs := make([]error, 0, retries)

	err := onRetry()

	for i := 0; i < retries-1 && err != nil; i++ {
		errs = append(errs, err)
		time.Sleep(sleep)
		err = onRetry()
	}

	if err != nil {
		errs = append(errs, err)
		return errs
	}

	return make([]error, 0)
}

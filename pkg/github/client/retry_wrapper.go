package ghclient

import (
	"errors"
	"fmt"
	"time"

	"github.com/arquillian/ike-prow-plugins/pkg/retry"
	gogh "github.com/google/go-github/github"
)

type retryWrapper struct {
	retries int
	sleep   time.Duration
}

// NewRetryWrapper creates an instance of retryWrapper that retries requests until either there is no error or limit is reached
func NewRetryWrapper(retries int, sleep time.Duration) AroundFunctionCreator {
	return &retryWrapper{
		retries: retries,
		sleep:   sleep,
	}
}

func (r retryWrapper) createAroundFunction(earlierAround aroundFunction) aroundFunction {
	return func(doFunction doFunction) doFunction {
		return func(aroundContext aroundContext) (func(), *gogh.Response, error) {
			return r.retry(earlierAround(doFunction), aroundContext)
		}
	}
}

func (r retryWrapper) retry(toRetry doFunction, aroundContext aroundContext) (func(), *gogh.Response, error) {
	var response *gogh.Response
	var setValueFunc func()
	errs := retry.Do(r.retries, r.sleep, func() error {
		var err error
		setValueFunc, response, err = toRetry(aroundContext)
		return err
	})

	if len(errs) == r.retries {
		msg := fmt.Sprintf("all %d attempts of sending a request failed. See the errors:", r.retries)
		for index, e := range errs {
			msg += fmt.Sprintf("\n%d. [%s]", index+1, e.Error())
		}
		return setValueFunc, response, errors.New(msg)
	}
	return setValueFunc, response, nil
}

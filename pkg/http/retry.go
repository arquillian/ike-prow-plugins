package http

import (
	gogh "github.com/google/go-github/github"
	"time"
	"fmt"
	"errors"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
)

// RequestInvoker is a function that invokes a request and returns a response
type RequestInvoker func() (*gogh.Response, error)

// Do invokes a request for the given amount of retries with the given sleep before each one of them until the response
// status code is lower than 404 or the returned error is not nil
func Do(retries int, sleep time.Duration, invokeRequest RequestInvoker) error {
	performedRetries := 0
	errs := utils.Retry(retries, sleep, func() (shouldRetry bool, error error) {
		response, err := invokeRequest()
		performedRetries++

		if response != nil && response.StatusCode >= 404 {
			if err != nil {
				return true, err
			}
			return true, fmt.Errorf("server responded with error %d", response.StatusCode)
		}
		return false, err
	})
	if len(errs) == performedRetries {
		msg := fmt.Sprintf("All %d attempts of sending a request failed. See the errors:", performedRetries)
		for index, e := range errs {
			msg = msg + fmt.Sprintf("\n%d. [%s]", index+1, e.Error())
		}
		return errors.New(msg)
	}
	return nil
}

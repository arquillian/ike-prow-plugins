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

// Do invokes a request for the given amount of retries with the given sleep before each one of them until the response status code is lower than 404
func Do(retries int, sleep time.Duration, invokeRequest RequestInvoker) error {
	errs := utils.Retry(retries, sleep, func() error {
		response, err := invokeRequest()

		if response != nil && response.StatusCode >= 404 {
			if err != nil {
				return err
			}
			return errors.New("Server responded with error " + string(response.StatusCode))
		}
		return nil
	})
	if len(errs) > 0 {
		msg := fmt.Sprintf("During %d retries of sending a request %d errors have occurred:", retries, len(errs))
		for index, e := range errs {
			msg = msg + fmt.Sprintf("\n%d. [%s]", index+1, e.Error())
		}
		return errors.New(msg)
	}
	return nil
}

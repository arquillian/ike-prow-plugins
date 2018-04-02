package http

import (
	gogh "github.com/google/go-github/github"
	"time"
	"fmt"
	"errors"
)

// RequestInvoker is a function that invokes a request and returns a response
type RequestInvoker func() (*gogh.Response, error)

// Do invokes a request for the given amount of retries with the given sleep before each one of them until the response status code is lower than 404
func Do(retries int, sleep time.Duration, invokeRequest RequestInvoker) error {
	errs := make([]error, 0, retries)

	response, err := invokeRequest()

	for i := 0; i < retries-1 && response != nil && response.StatusCode >= 404; i++ {
		errs = append(errs, err)
		time.Sleep(sleep)
		response, err = invokeRequest()
	}

	if err != nil {
		errs = append(errs, err)
		msg := fmt.Sprintf("During %d retries of sending a request %d errs have occurred:", retries, len(errs))
		for index, e := range errs {
			msg = msg + fmt.Sprintf("\n%d. [%s]", index + 1, e.Error())
		}
		return errors.New(msg)
	}
	return nil
}

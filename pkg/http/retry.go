package http

import (
	gogh "github.com/google/go-github/github"
	"time"
)

// RequestInvoker is a function that invokes a request and returns a response
type RequestInvoker func() *gogh.Response

// Do invokes a request for the given amount of retries with the given sleep before each one of them until the response status code is lower than 404
func Do(retries int, sleep time.Duration, invokeRequest RequestInvoker) {
	response := invokeRequest()
	for i := 0; i < retries-1 && response != nil && response.StatusCode >= 404; i++ {
		time.Sleep(sleep)
		response = invokeRequest()
	}
}

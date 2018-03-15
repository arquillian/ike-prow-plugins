package utils

import (
	"errors"
	"io/ioutil"
	"net/http"
)

// GetFileFromURL retrieves the content of the file on the given url
func GetFileFromURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return []byte(""), errors.New("The response's status code wasn't 2xx, but " + string(resp.StatusCode))
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

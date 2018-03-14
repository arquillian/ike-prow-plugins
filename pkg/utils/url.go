package utils

import (
	"io/ioutil"
	"net/http"
)

// GetFileFromURL retrieves the content of the file on the given url
func GetFileFromURL(url string) ([]byte, bool, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, false, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return []byte(""), false, nil
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, false, err
	}

	return body, true, nil
}

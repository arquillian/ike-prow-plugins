package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// GetFileFromURL retrieves the content of the file on the given url
func GetFileFromURL(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return make([]byte, 0), fmt.Errorf("server responded with error %d", resp.StatusCode)
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

package utils

import (
	"bytes"
	"io/ioutil"
)

// LoadSecret reads bytes from the file
func LoadSecret(secretFilename string) ([]byte, error) {
	rawSecret, err := ioutil.ReadFile(secretFilename)
	if err != nil {
		return nil, err
	}
	return bytes.TrimSpace(rawSecret), nil
}

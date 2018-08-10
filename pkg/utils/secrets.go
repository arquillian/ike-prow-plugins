package utils

import (
	"bytes"
	"io/ioutil"
)

// LoadSecret reads bytes from the file
func LoadSecret(secretFilename string) ([]byte, error) {
	// This is only executed from within a container while starting up the process
	rawSecret, err := ioutil.ReadFile(secretFilename) // nolint
	if err != nil {
		return nil, err
	}
	return bytes.TrimSpace(rawSecret), nil
}

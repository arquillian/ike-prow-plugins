package utils

import (
	"bytes"
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

var (
	log = logrus.StandardLogger().WithField("ike-plugins", "secrets-loader")
)

// LoadSecret reads bytes from the file
func LoadSecret(secretFilename string) []byte {
	rawSecret, err := ioutil.ReadFile(secretFilename)
	if err != nil {
		log.WithError(err).Fatalf("Could not read %q secret file.", secretFilename)
	}
	return bytes.TrimSpace(rawSecret)
}

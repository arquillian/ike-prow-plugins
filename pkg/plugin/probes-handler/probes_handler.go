package probeshandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

// Probe defines the json elements for the probesHandler response body content
type Probe struct {
	Version string
}

// NewProbesHandler registers liveliness and readinesss probes for /version endpoint
func NewProbesHandler(log *logrus.Entry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		serverError := func(action string, err error) {
			log.WithError(err).Errorf("Error %s.", action)
			msg := fmt.Sprintf("500 Internal server error %s: %v", action, err)
			http.Error(w, msg, http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		content, err := getJSONContent()
		if err != nil {
			serverError("marshaling response body: %v", err)
		}
		if _, err := w.Write(content); err != nil {
			serverError("writing response body: %v", err)
		}
	})
}

func getJSONContent() ([]byte, error) {
	version, found := os.LookupEnv("VERSION")
	if !found {
		version = "UNKNOWN"
	}
	js, err := json.Marshal(Probe{Version: version})
	if err != nil {
		return nil, err
	}
	return js, nil
}

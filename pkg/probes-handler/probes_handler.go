package probeshandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/arquillian/ike-prow-plugins/pkg/log"
)

// Probe defines the json elements for the probesHandler response body content
type Probe struct {
	Version string
}

// NewProbesHandler registers liveliness and readiness probes for /version endpoint
func NewProbesHandler(logger log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {

		serverError := func(action string, err error) {
			msg := fmt.Sprintf("Probe handler failed while %s: %v.", action, err)
			logger.Errorf(msg)
			msg = fmt.Sprintf("%d %s:\n%s",
				http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), msg)
			http.Error(w, msg, http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		content, err := getJSONContent()
		if err != nil {
			serverError("marshaling response body", err)
		}
		if _, err := w.Write(content); err != nil {
			serverError("writing response body", err)
		}
	})
}

func getJSONContent() ([]byte, error) {
	version, found := os.LookupEnv("VERSION")
	if !found {
		version = "UNKNOWN"
	}
	result, err := json.Marshal(Probe{Version: version})
	if err != nil {
		return nil, err
	}
	return result, nil
}

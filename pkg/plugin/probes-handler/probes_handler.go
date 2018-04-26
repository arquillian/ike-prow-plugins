package probeshandler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// Probe defines the json elements for the probesHandler response body content
type Probe struct {
	Version string
}

// NewProbesHandler registers liveliness and readinesss probes for /version endpoint
func NewProbesHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write(getJSONContent())
		if err != nil {
			log.Printf("Writing failed: %v", err)
		}
	})
}

func getJSONContent() []byte {
	version, found := os.LookupEnv("VERSION")
	if !found {
		version = "UNKNOWN"
	}
	js, err := json.Marshal(Probe{version})
	if err != nil {
		log.Printf("Marshalling failed: %v", err)
	}
	return js
}

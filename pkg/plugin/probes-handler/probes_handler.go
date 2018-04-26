package probeshandler

import (
	"log"
	"net/http"
	"os"
)

// NewProbesHandler registers liveliness and readinesss probes for /version endpoint
func NewProbesHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(os.Getenv("VERSION")))
		if err != nil {
			log.Printf("Write failed: %v", err)
		}
	})
}

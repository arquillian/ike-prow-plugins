package plugin

import (
	"log"
	"net/http"
)

// NewProbesHandler registers liveliness and readinesss probes for /version endpoint
func NewProbesHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("200"))
		if err != nil {
			log.Printf("Write failed: %v", err)
		}
	})
}

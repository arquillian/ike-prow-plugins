package probeshandler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/arquillian/ike-prow-plugins/pkg/log"
	probeshandler "github.com/arquillian/ike-prow-plugins/pkg/probes-handler"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test liveliness and readiness probes.", func() {

	const (
		probesEndpoint = "/version"
		defaultVersion = "xxxxxxxx-xxxxxxxxxx"
	)

	var versionEnv string

	var _ = BeforeSuite(func() {
		var found bool
		if versionEnv, found = os.LookupEnv("VERSION"); !found {
			os.Setenv("VERSION", defaultVersion)
		}
	})

	var _ = AfterSuite(func() {
		os.Setenv("VERSION", versionEnv)
	})

	Context("When in healthy state", func() {
		It("should return plugin version in response body", func() {
			// given
			probesHandler := probeshandler.NewProbesHandler(log.NewTestLogger())
			request := httptest.NewRequest("GET", probesEndpoint, nil)
			response := httptest.NewRecorder()
			expectedBody := probeshandler.Probe{Version: os.Getenv("VERSION")}

			// when
			http.Handle(probesEndpoint, probesHandler)
			probesHandler.ServeHTTP(response, request)

			// then
			actualBody := probeshandler.Probe{}
			json.Unmarshal(response.Body.Bytes(), &actualBody)
			Expect(actualBody).To(Equal(expectedBody))
		})
	})
})

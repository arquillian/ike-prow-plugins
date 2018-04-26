package probeshandler_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"

	probeshandler "github.com/arquillian/ike-prow-plugins/pkg/plugin/probes-handler"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test liveliness and readiness probes.", func() {

	const (
		probesEndpoint = "/version"
		version        = "xxxxxxxx-xxxxxxxxxx"
	)

	var _ = BeforeSuite(func() {
		os.Setenv("VERSION", version)
	})

	var _ = AfterSuite(func() {
		os.Clearenv()
	})

	Context("When in healthy state", func() {
		It("should return plugin version in response body", func() {
			// given
			probesHandler := probeshandler.NewProbesHandler()
			request := httptest.NewRequest("GET", probesEndpoint, nil)
			response := httptest.NewRecorder()

			// when
			http.Handle(probesEndpoint, probesHandler)
			probesHandler.ServeHTTP(response, request)

			// then
			responseBody, _ := ioutil.ReadAll(response.Result().Body)
			Expect(string(responseBody)).To(Equal("version: " + version))
		})
	})

	Context("When in unhealthy state", func() {
		It("should return empty response body", func() {
			// given
			handler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
			}
			request := httptest.NewRequest("GET", probesEndpoint, nil)
			response := httptest.NewRecorder()

			// when
			handler(response, request)

			// then
			responseBody, _ := ioutil.ReadAll(response.Result().Body)
			Expect(string(responseBody)).To(Equal(""))
		})
	})
})

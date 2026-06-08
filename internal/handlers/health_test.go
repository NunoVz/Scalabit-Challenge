package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HealthHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Incorrect Status: got %v, expected %v", status, http.StatusOK)
	}

	if rr.Body.String() != "OK" {
		t.Errorf("Incorrect Body: got %v, expected OK", rr.Body.String())
	}
}

type mockErrorResponseWriter struct{}

func (m *mockErrorResponseWriter) Header() http.Header { return http.Header{} }
func (m *mockErrorResponseWriter) Write([]byte) (int, error) {
	return 0, errors.New("simulated write error")
}
func (m *mockErrorResponseWriter) WriteHeader(statusCode int) {}

func TestHealthHandler_WriteError(t *testing.T) {
	req, _ := http.NewRequest("GET", "/health", nil)
	w := &mockErrorResponseWriter{}

	HealthHandler(w, req)
}

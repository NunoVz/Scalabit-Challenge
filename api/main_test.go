package main

import (
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
	handler := http.HandlerFunc(healthHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Incorrect Status: got %v, expected %v", status, http.StatusOK)
	}

	if rr.Body.String() != "OK" {
		t.Errorf("Incorrect Body: got %v, expected OK", rr.Body.String())
	}
}

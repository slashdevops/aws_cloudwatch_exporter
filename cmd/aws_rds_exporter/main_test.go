package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/healthz", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthHandler)
	handler.ServeHTTP(rr, req)

	status := rr.Code
	if status != http.StatusOK {
		t.Errorf("Handler returned %v", status)
	}

	// 	expected := `OK`
	// 	if rr.Body.String() != expected {
	// 		t.Errorf("Handler returned %v", rr.Body.String())
	// 	}
}

func TestHomeHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(homeHandler)
	handler.ServeHTTP(rr, req)

	status := rr.Code
	if status != http.StatusOK {
		t.Errorf("Handler returned %v", status)
	}
}

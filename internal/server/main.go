package server

import (
	"net/http"
	"time"
)

func New(mux *http.ServeMux, addr string) *http.Server {
	return &http.Server{
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		Addr:              addr,
		Handler:           mux,
	}
}
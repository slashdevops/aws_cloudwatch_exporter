package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/slashdevops/aws_cloudwatch_exporter/config"
)

func New(mux *http.ServeMux, c *config.All) *http.Server {
	return &http.Server{
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		Addr:              c.Server.Address + strconv.Itoa(int(c.Server.Port)),
		Handler:           mux,
	}
}

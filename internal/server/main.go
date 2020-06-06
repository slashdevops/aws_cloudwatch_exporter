package server

import (
	"net/http"
	"strconv"

	"github.com/slashdevops/aws_cloudwatch_exporter/config"
)

func New(mux *http.ServeMux, c *config.All) *http.Server {
	s := &http.Server{
		ReadTimeout:       c.Server.ReadTimeout,
		WriteTimeout:      c.Server.WriteTimeout,
		IdleTimeout:       c.Server.IdleTimeout,
		ReadHeaderTimeout: c.Server.ReadHeaderTimeout,
		Addr:              c.Server.Address + ":" + strconv.Itoa(int(c.Server.Port)),
		Handler:           mux,
	}

	s.SetKeepAlivesEnabled(c.Server.KeepAlivesEnabled)
	return s
}

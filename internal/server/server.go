package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/slashdevops/aws_cloudwatch_exporter/config"
)

type Server struct {
	c *config.All
	s *http.Server
}

func New(mux *http.ServeMux, c *config.All) *Server {
	httpServer := &http.Server{
		ReadTimeout:       c.Server.ReadTimeout,
		WriteTimeout:      c.Server.WriteTimeout,
		IdleTimeout:       c.Server.IdleTimeout,
		ReadHeaderTimeout: c.Server.ReadHeaderTimeout,
		Addr:              c.Server.Address + ":" + strconv.Itoa(int(c.Server.Port)),
		Handler:           mux,
	}

	httpServer.SetKeepAlivesEnabled(c.Server.KeepAlivesEnabled)

	server := &Server{
		c: c,
		s: httpServer,
	}
	return server
}

func (s *Server) ListenOSSignals(done *chan bool) {
	go func(s *Server, done *chan bool) {
		osSignals := make(chan os.Signal, 1)
		signal.Notify(osSignals, os.Interrupt)
		signal.Notify(osSignals, syscall.SIGTERM)
		signal.Notify(osSignals, syscall.SIGINT)
		signal.Notify(osSignals, syscall.SIGQUIT)

		log.Info("Listen Operating System signals")
		sig := <-osSignals
		log.Infof("Received signal %s from operation system", sig)
		s.doGracefullyShutdown()

		// Notify main routine shutdown is done
		*done <- true
	}(s, done)
}

func (s *Server) doGracefullyShutdown() {
	log.Infof("Graceful shutdown, wait %vs\n", s.c.Server.ShutdownTimeout.Seconds())

	ctx, cancel := context.WithTimeout(context.Background(), s.c.ShutdownTimeout)
	defer cancel()

	s.s.SetKeepAlivesEnabled(false)

	if err := s.s.Shutdown(ctx); err != nil {
		log.Fatalf("Server was shutdown, %s\n", err)
	}
	log.Info("Server stopped")
}

func (s *Server) Start() (err error) {
	log.Info("Server starting")
	if err := s.s.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return
}

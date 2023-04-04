/*
Copyright © 2020 Christian González Di Antonio christian@slashdevops.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/slashdevops/aws_cloudwatch_exporter/internal/config"
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

		log.Info("Server is listening Operating System signals")
		sig := <-osSignals
		log.Warnf("Received signal %s from Operation System", sig)
		s.doGracefullyShutdown()

		// Notify main routine shutdown is done
		*done <- true
	}(s, done)
}

func (s *Server) doGracefullyShutdown() {
	log.Warnf("Graceful shutdown, wait at least %vs before stop\n", s.c.Server.ShutdownTimeout.Seconds())

	ctx, cancel := context.WithTimeout(context.Background(), s.c.ShutdownTimeout)
	defer cancel()

	s.s.SetKeepAlivesEnabled(false)

	if err := s.s.Shutdown(ctx); err != nil {
		log.Fatalf("Server was shutdown, %s\n", err)
	}
	log.Info("Server stopped")
}

func (s *Server) Start() (err error) {
	log.Infof("Server starting on %s:%v", s.c.Server.Address, s.c.Server.Port)
	if err := s.s.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return
}

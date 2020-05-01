package web

import (
	"log"
	"net/http"
	"net/http/pprof"

	"github.com/aimroot/aws_cloudwatch_exporter/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"text/template"
)

type Handlers struct {
	conf   *config.Server
	logger *log.Logger
}

func NewHandlers(l *log.Logger, c *config.Server) *Handlers {
	return &Handlers{
		logger: l,
		conf:   c,
	}
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title         string
		MetricHandler string
	}{h.conf.App.Description, h.conf.Server.MetricsPath}
	t := template.Must(template.ParseFiles("web/templates/index.html"))
	t.Execute(w, data)
}

func (h *Handlers) health(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
}

func (h *Handlers) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", h.Home)
	mux.HandleFunc("/healthz", h.health)

	// Prometheus endopoint
	mux.Handle(h.conf.Server.MetricsPath, promhttp.Handler())

	// Debug & Profiling
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

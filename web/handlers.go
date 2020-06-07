package web

import (
	"net/http"

	"github.com/google/martian/log"
	"github.com/slashdevops/aws_cloudwatch_exporter/config"

	"text/template"
)

type Handlers struct {
	conf *config.All
}

func NewHandlers(c *config.All) *Handlers {
	return &Handlers{
		conf: c,
	}
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	indexHtmlTmpl := `
<html>
<head>
    <title>{{.Title}}</title>
</head>
<body>
    <h1><a href="{{.GitRepository}}">{{.Name}}</a></h1>
	<h2>{{.Description}}</h2>
	<h3>Links:</h3>
	<ul>
		<li><a href="{{.MetricsPath}}">{{.MetricsPath}}</a></li>
		<li><a href="{{.HealthPath}}">{{.HealthPath}}</a></li>
	</ul>

	<h2>Version</h2>
	<ul>
		<li>{{.VersionInfo}}</li>
		<li>{{.BuildInfo}}</li>
	</ul>
	
	{{ if .Debug }}
	<h2>Debug is enabled</h2>
	<h3>Links:</h3>
	<ul>
		{{range .DebugLinks}}
		<li><a href="{{.}}">{{.}}</a></li>
		{{ end }}
	</ul>
	{{ end }}

	<h3><a href="https://prometheus.io/">https://prometheus.io</a></h3>
</body>
</html>
`
	data := struct {
		Title         string
		Name          string
		Description   string
		GitRepository string
		MetricsPath   string
		HealthPath    string
		VersionInfo   string
		BuildInfo     string
		Debug         bool
		DebugLinks    []string
	}{
		h.conf.Application.Name,
		h.conf.Application.Name,
		h.conf.Application.Description,
		h.conf.Application.GitRepository,
		h.conf.Application.MetricsPath,
		h.conf.Application.HealthPath,
		h.conf.Application.VersionInfo,
		h.conf.Application.BuildInfo,
		h.conf.Server.Debug,
		[]string{
			"/debug/pprof/",
			"/debug/pprof/cmdline",
			"/debug/pprof/profile",
			"/debug/pprof/symbol",
			"/debug/pprof/trace",
		},
	}

	t := template.Must(template.New("index").Parse(indexHtmlTmpl))
	if err := t.Execute(w, data); err != nil {
		log.Errorf("Error rendering template %s", err)
	}
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
}

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
package web

import (
	"net/http"

	log "github.com/sirupsen/logrus"
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
	indexHTMLTmpl := `
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
	{{ else }}
	<h2>Debug is disabled, you cannot see application performance profile</h2>
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

	t := template.Must(template.New("index").Parse(indexHTMLTmpl))
	if err := t.Execute(w, data); err != nil {
		log.Errorf("Error rendering template: %s", err)
	}
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
}

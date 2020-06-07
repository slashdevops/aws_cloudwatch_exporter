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
    <h1><a href="{{.MetricHandler}}">metrics</a></h1>
</body>
</html>
`
	data := struct {
		Title         string
		MetricHandler string
	}{h.conf.Application.Description, h.conf.Application.MetricsPath}

	t := template.Must(template.New("index").Parse(indexHtmlTmpl))
	if err := t.Execute(w, data); err != nil {
		log.Errorf("Error rendering template %s", err)
	}
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
}

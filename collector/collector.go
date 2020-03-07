package collector

import (
	"io"
	"log"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricRule struct {
	Namespace                    string
	MetricName                   string
	Period                       int32
	Range                        int32
	delay                        int32
	Statistics                   []string
	StatistExtendedStatisticsics []string
	Dimensions                   []string
	DimensionSelect              map[string]string
	DimensionSelectRegex         map[string]string
	Help                         string
	TimeStamp                    bool
}

var rules []MetricRule

type Exporter struct {
	URI   string
	mutex sync.RWMutex
	fetch func() (io.ReadCloser, error)

	up                             prometheus.Gauge
	totalScrapes, csvParseFailures prometheus.Counter
	serverMetrics                  map[int]*prometheus.Desc
	logger                         log.Logger
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {

}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
}

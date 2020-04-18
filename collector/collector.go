package collector

import (
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
)

// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/aws-services-cloudwatch-metrics.html
var (
	up = prometheus.NewDesc(
		"up",
		"Was talking to aws_cloudwatch_exporter successful.",
		nil, nil,
	)
	invalidChars = regexp.MustCompile("[^a-zA-Z0-9:_]")
)

type CWCollector struct {
	CloudWatchClient string
	Scrapes          prometheus.Counter
	//Collectors       []collector
}

func (c *CWCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
}

func (c *CWCollector) Collect(ch chan<- prometheus.Metric) {
}

package collector

import (
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/slashdevops/aws_cloudwatch_exporter/config"
)

// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/aws-services-cloudwatch-metrics.html
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/using-metric-math.html#metric-math-syntax
// https://docs.aws.amazon.com/sdk-for-go/api/service/cloudwatch/#CloudWatch.GetMetricData
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html
// https://aws.amazon.com/premiumsupport/knowledge-center/cloudwatch-getmetricdata-api/
var (
	up = prometheus.NewDesc(
		"up",
		"aws_cloudwatch_exporter is up and running.",
		nil, nil,
	)
	invalidChars = regexp.MustCompile("[^a-zA-Z0-9:_]")
)

type AWSCollector struct {
	conf    *config.All
	Up      *prometheus.Desc
	Scrapes prometheus.Counter
}

func NewAWSCollector(c *config.All) *AWSCollector {
	return &AWSCollector{
		conf: c,
		Up: prometheus.NewDesc(
			c.Application.Namespace+"_up",
			c.Application.Name+" is up and running",
			nil,
			nil,
		),
		Scrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: c.Application.Namespace,
			Subsystem: "collector",
			Name:      "scrapes_total",
			Help:      "Total number of times AWS CloudWatch API was scraped for metrics.",
		}),
	}
}

// Implements prometheus.Collector
func (c *AWSCollector) Describe(ch chan<- *prometheus.Desc) {
	c.Scrapes.Describe(ch)
	ch <- up
}

// Implements prometheus.Collector
func (c *AWSCollector) Collect(ch chan<- prometheus.Metric) {
	c.Scrapes.Inc()
	// When the collector is working fine
	ch <- prometheus.MustNewConstMetric(c.Up, prometheus.GaugeValue, 1)
}

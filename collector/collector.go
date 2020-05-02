package collector

import (
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
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

type ACWCollector struct {
	CloudWatchClient string
	Scrapes          prometheus.Counter
	//Collectors       []collector
}

// Implements prometheus.Collector
func (c *ACWCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
}

// Implements prometheus.Collector
func (c *ACWCollector) Collect(ch chan<- prometheus.Metric) {

	// When the collector is working fine
	ch <- prometheus.MustNewConstMetric(up, prometheus.GaugeValue, 1)
}

package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/slashdevops/aws_cloudwatch_exporter/config"
)

// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/aws-services-cloudwatch-metrics.html
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/using-metric-math.html#metric-math-syntax
// https://docs.aws.amazon.com/sdk-for-go/api/service/cloudwatch/#CloudWatch.GetMetricData
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html
// https://aws.amazon.com/premiumsupport/knowledge-center/cloudwatch-getmetricdata-api/
// https://aws.amazon.com/cloudwatch/pricing/
type Collector struct {
	conf    *config.All
	Up      prometheus.Gauge
	Info    prometheus.Gauge
	Scrapes prometheus.Counter
}

func New(c *config.All) *Collector {
	return &Collector{
		conf: c,
		Up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: c.Application.Namespace + "_up",
			Subsystem: "collector",
			Name:      c.Application.Name + "_up",
			Help:      "Was the last scrape of " + c.Application.Name + " successful.",
		}),
		Info: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace:   c.Application.Namespace,
				Subsystem:   "collector",
				Name:        c.Application.Name + "_version_info",
				Help:        c.Application.Name + " version info.",
				ConstLabels: prometheus.Labels{"release_date": c.BuildInfo, "version": c.Version},
			},
		),
		Scrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: c.Application.Namespace,
			Subsystem: "collector",
			Name:      c.Application.Name + "_scrapes_total",
			Help:      "Total number of times AWS CloudWatch API was scraped for metrics.",
		}),
	}
}

// Implements prometheus.Collector Interface
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.Scrapes.Describe(ch)
	c.Info.Describe(ch)
	c.Up.Describe(ch)
}

// Implements prometheus.Collector Interface
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.Info.Set(1)
	ch <- c.Info

	c.Scrapes.Inc()
	ch <- c.Scrapes

	// When the collector is working fine
	c.Up.Set(1)
	ch <- c.Up
}

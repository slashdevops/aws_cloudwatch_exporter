package collector

import (
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/slashdevops/aws_cloudwatch_exporter/config"
	"github.com/slashdevops/aws_cloudwatch_exporter/internal/awshelper"
	"github.com/slashdevops/aws_cloudwatch_exporter/internal/metrics"
)

// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/aws-services-cloudwatch-metrics.html
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/using-metric-math.html#metric-math-syntax
// https://docs.aws.amazon.com/sdk-for-go/api/service/cloudwatch/#CloudWatch.GetMetricData
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html
// https://aws.amazon.com/premiumsupport/knowledge-center/cloudwatch-getmetricdata-api/
// https://aws.amazon.com/cloudwatch/pricing/

const (
	SUCCESS = "success"
	ERROR   = "error"
)

type OwnMetrics struct {
	Up                    prometheus.Gauge
	Info                  prometheus.Gauge
	ScrapesSuccess        prometheus.Counter
	ScrapesErrors         prometheus.Counter
	MetricsScrapesSuccess prometheus.Counter
	MetricsScrapesErrors  prometheus.Counter
}

type Collector struct {
	conf       *config.All
	metrics    metrics.Metrics
	mutex      sync.RWMutex
	ownMetrics *OwnMetrics
}

func New(c *config.All, m metrics.Metrics) *Collector {
	return &Collector{
		conf:    c,
		metrics: m,
		ownMetrics: &OwnMetrics{
			Up: prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace: c.Application.Namespace,
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
			ScrapesSuccess: prometheus.NewCounter(
				prometheus.CounterOpts{
					Namespace:   c.Application.Namespace,
					Subsystem:   "collector",
					Name:        c.Application.Name + "_scrapes_success_total",
					Help:        "Total number of times of AWS CloudWatch API was scraped for metrics with success result.",
					ConstLabels: nil,
				},
			),
			ScrapesErrors: prometheus.NewCounter(
				prometheus.CounterOpts{
					Namespace:   c.Application.Namespace,
					Subsystem:   "collector",
					Name:        c.Application.Name + "_scrapes_errors_total",
					Help:        "Total number of times of AWS CloudWatch API was scraped for metrics with error result.",
					ConstLabels: nil,
				},
			),
			MetricsScrapesSuccess: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   c.Application.Namespace,
					Subsystem:   "collector",
					Name:        c.Application.Name + "_metrics_scrapes_success_total",
					Help:        "Total number of metrics of AWS CloudWatch API was scraped with success result.",
					ConstLabels: nil,
				},
			),
			MetricsScrapesErrors: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   c.Application.Namespace,
					Subsystem:   "collector",
					Name:        c.Application.Name + "_metrics_scrapes_errors_total",
					Help:        "Total number of metrics of AWS CloudWatch API was scraped with errors result.",
					ConstLabels: nil,
				},
			),
		},
	}
}

// Implements prometheus.Collector Interface
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.ownMetrics.Info.Desc()
	ch <- c.ownMetrics.Up.Desc()
	c.ownMetrics.ScrapesSuccess.Describe(ch)
	c.ownMetrics.ScrapesErrors.Describe(ch)
	c.ownMetrics.MetricsScrapesSuccess.Describe(ch)
	c.ownMetrics.MetricsScrapesErrors.Describe(ch)

	// Describe all metrics created from yaml files
	for _, md := range c.metrics.GetMetricsDesc() {
		ch <- md
	}
}

// Implements prometheus.Collector Interface
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock() // To protect metrics from concurrent collects.
	defer c.mutex.Unlock()

	c.ownMetrics.Info.Set(1)
	ch <- c.ownMetrics.Info

	c.scrape(ch)
}

func (c *Collector) scrape(ch chan<- prometheus.Metric) {
	c.ownMetrics.Up.Set(1)

	startTime, endTime, period := metrics.GetTimeStamps(time.Now(), c.conf.Application.StatsPeriod)
	mdi := c.metrics.GetMetricDataInput(startTime, endTime, period, "")

	sess, _ := awshelper.NewSession(&c.conf.AWS)
	svc := cloudwatch.New(sess)

	// Scrape CloudWatch Metrics
	mdo, err := svc.GetMetricData(mdi)
	if err != nil {
		c.ownMetrics.Up.Set(0)
		c.ownMetrics.ScrapesSuccess.Inc()
		log.Errorf("Error getting metrics %v", err)
	}
	c.ownMetrics.ScrapesSuccess.Inc()

	for _, mdr := range mdo.MetricDataResults {
		ch <- c.ownMetrics.MetricsScrapesSuccess

		for i, t := range mdr.Timestamps {
			nm := prometheus.NewMetricWithTimestamp(
				*t,
				prometheus.MustNewConstMetric(
					c.metrics.GetMetricDesc(*mdr.Id),
					prometheus.GaugeValue,
					*mdr.Values[i],
				),
			)

			c.metrics.SetMetric(*mdr.Id, nm)
		}
	}

	// report own metrics
	ch <- c.ownMetrics.ScrapesSuccess
	ch <- c.ownMetrics.ScrapesErrors
	ch <- c.ownMetrics.MetricsScrapesSuccess
	ch <- c.ownMetrics.MetricsScrapesErrors
	ch <- c.ownMetrics.Up
}

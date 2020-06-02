package collector

import (
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/slashdevops/aws_cloudwatch_exporter/config"
	"github.com/slashdevops/aws_cloudwatch_exporter/internal/metrics"
)

// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/aws-services-cloudwatch-metrics.html
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/using-metric-math.html#metric-math-syntax
// https://docs.aws.amazon.com/sdk-for-go/api/service/cloudwatch/#CloudWatch.GetMetricData
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html
// https://aws.amazon.com/premiumsupport/knowledge-center/cloudwatch-getmetricdata-api/
// https://aws.amazon.com/cloudwatch/pricing/

type OwnMetrics struct {
	Up                     prometheus.Gauge
	Info                   prometheus.Gauge
	ScrapesSuccess         prometheus.Counter
	ScrapesErrors          prometheus.Counter
	ScrapesMessages        prometheus.Counter
	MetricsScrapesSuccess  prometheus.Counter
	MetricsScrapesErrors   prometheus.Counter
	MetricsScrapesEmpty    prometheus.Counter
	MetricsScrapesMessages prometheus.Counter
}

type Collector struct {
	conf       *config.All
	sess       *session.Session
	metrics    metrics.Metrics
	mutex      sync.RWMutex
	ownMetrics *OwnMetrics
}

func New(c *config.All, m metrics.Metrics, s *session.Session) *Collector {
	return &Collector{
		conf:    c,
		sess:    s,
		metrics: m,
		ownMetrics: &OwnMetrics{
			Up: prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace: c.Application.Namespace,
				Subsystem: "collector",
				Name:      "up",
				Help:      "Was the last scrape of " + c.Application.Name + " successful.",
			}),
			Info: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   c.Application.Namespace,
					Subsystem:   "collector",
					Name:        "version_info",
					Help:        c.Application.Name + " version info.",
					ConstLabels: prometheus.Labels{"release_date": c.BuildInfo, "version": c.Version},
				},
			),
			ScrapesSuccess: prometheus.NewCounter(
				prometheus.CounterOpts{
					Namespace:   c.Application.Namespace,
					Subsystem:   "collector",
					Name:        "scrapes_success_total",
					Help:        "Total number of times of AWS CloudWatch API was scraped for metrics with success result.",
					ConstLabels: nil,
				},
			),
			ScrapesErrors: prometheus.NewCounter(
				prometheus.CounterOpts{
					Namespace:   c.Application.Namespace,
					Subsystem:   "collector",
					Name:        "scrapes_errors_total",
					Help:        "Total number of times of AWS CloudWatch API was scraped for metrics with error result.",
					ConstLabels: nil,
				},
			),
			ScrapesMessages: prometheus.NewCounter(
				prometheus.CounterOpts{
					Namespace:   c.Application.Namespace,
					Subsystem:   "collector",
					Name:        "scrapes_messages_total",
					Help:        "Total number of times of AWS CloudWatch API was scraped for metrics with some message result. (see the logs)",
					ConstLabels: nil,
				},
			),
			MetricsScrapesSuccess: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   c.Application.Namespace,
					Subsystem:   "collector",
					Name:        "metrics_scrapes_success_total",
					Help:        "Total number of metrics of AWS CloudWatch API was scraped with success result.",
					ConstLabels: nil,
				},
			),
			MetricsScrapesErrors: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   c.Application.Namespace,
					Subsystem:   "collector",
					Name:        "metrics_scrapes_errors_total",
					Help:        "Total number of metrics of AWS CloudWatch API was scraped with errors result.",
					ConstLabels: nil,
				},
			),
			MetricsScrapesEmpty: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   c.Application.Namespace,
					Subsystem:   "collector",
					Name:        "metrics_scrapes_empty_total",
					Help:        "Total number of metrics of AWS CloudWatch API was scraped with empty result.",
					ConstLabels: nil,
				},
			),
			MetricsScrapesMessages: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   c.Application.Namespace,
					Subsystem:   "collector",
					Name:        "metrics_scrapes_messages_total",
					Help:        "Total number of metrics of AWS CloudWatch API was scraped with some messages result. (see the logs)",
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
	c.ownMetrics.ScrapesMessages.Describe(ch)
	c.ownMetrics.MetricsScrapesSuccess.Describe(ch)
	c.ownMetrics.MetricsScrapesErrors.Describe(ch)
	c.ownMetrics.MetricsScrapesEmpty.Describe(ch)
	c.ownMetrics.MetricsScrapesMessages.Describe(ch)

	// Describe all metrics created from yaml files
	for _, md := range c.metrics.GetMetricsDesc() {
		ch <- md
	}

	/*	for _, md := range c.metrics.GetMetrics() {
		ch <- md.Desc()
	}*/
}

// Implements prometheus.Collector Interface
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock() // To protect metrics from concurrent collects.
	defer c.mutex.Unlock()

	// this information is constant
	c.ownMetrics.Info.Set(1)
	ch <- c.ownMetrics.Info

	c.scrape(ch)
}

// this do the job of scrape the metrics, parse the response from AWS CloudWatch and
// create the prometheus metrics
func (c *Collector) scrape(ch chan<- prometheus.Metric) {
	c.ownMetrics.Up.Set(1)

	// get the timestamps necessary to query metrics from AWS CloudWatch
	//              points     period        now()
	//                ↓        ↓→  ←↓         ↓
	// [(startTime)............................(endTime)] → time
	startTime, endTime, period := metrics.GetTimeStamps(
		time.Now(),
		c.conf.Application.MetricStatPeriod,
		c.conf.Application.MetricTimeWindow)

	mdi := c.metrics.GetMetricDataInput(startTime, endTime, period, "")

	svc := cloudwatch.New(c.sess)

	// Scrape CloudWatch Metrics
	mdo, err := svc.GetMetricData(mdi)
	if err != nil {
		c.ownMetrics.Up.Set(0)
		c.ownMetrics.ScrapesErrors.Inc()
		log.Errorf("Error getting metrics %v", err)
	}
	c.ownMetrics.ScrapesSuccess.Inc()

	// Some information came from the metrics scrape
	if len(mdo.Messages) > 0 {
		c.ownMetrics.ScrapesMessages.Inc()
		var msgs []string
		for _, m := range mdo.Messages {
			msgs = append(msgs, *m.Value)
		}
		mgssString := strings.Join(msgs, ",")
		log.Warnf("GetMetricDataOutput Message field contain: %s", mgssString)
	}

	for _, mdr := range mdo.MetricDataResults {

		if *mdr.StatusCode == "InternalError" {
			c.ownMetrics.MetricsScrapesErrors.Inc()
			log.Errorf("Error gotten when scrap metric id: %s, label: %s", *mdr.Id, *mdr.Label)
			continue
		}

		// Some information came from the metric scrape
		if len(mdr.Messages) > 0 {
			c.ownMetrics.MetricsScrapesMessages.Inc()
			var msgs []string
			for _, m := range mdr.Messages {
				msgs = append(msgs, *m.Value)
			}
			mgssString := strings.Join(msgs, ",")
			log.Warnf("Message field for metric id: %s, contain: %s", *mdr.Id, mgssString)
		}

		// no metric value came, continue with the next
		if len(mdr.Values) == 0 {
			c.ownMetrics.MetricsScrapesEmpty.Inc()
			continue
		}

		// mdr.Timestamps[0] and mdr.Values[0] because the first value into de arrays is the newest value
		// since we set ScanBy: TimestampDescending into GetMetricDataInput()
		nm := prometheus.NewMetricWithTimestamp(
			*mdr.Timestamps[0],
			prometheus.MustNewConstMetric(
				c.metrics.GetMetricDesc(*mdr.Id),
				prometheus.GaugeValue,
				*mdr.Values[0],
			),
		)

		//c.metrics.SetMetric(*mdr.Id, nm)

		c.ownMetrics.MetricsScrapesSuccess.Inc()

		// Send metric to channel
		ch <- nm
	}

	// report own metrics
	ch <- c.ownMetrics.Up
	ch <- c.ownMetrics.ScrapesSuccess
	ch <- c.ownMetrics.ScrapesErrors
	ch <- c.ownMetrics.ScrapesMessages
	ch <- c.ownMetrics.MetricsScrapesSuccess
	ch <- c.ownMetrics.MetricsScrapesErrors
	ch <- c.ownMetrics.MetricsScrapesEmpty
	ch <- c.ownMetrics.MetricsScrapesMessages
}

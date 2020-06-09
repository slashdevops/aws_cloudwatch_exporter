package collector

import (
	"fmt"
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
	MetricsTotal           prometheus.Gauge
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
				Name:      "up",
				Help:      "Was the last scrape of " + c.Application.Name + " successful.",
			}),
			Info: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace: c.Application.Namespace,
					Name:      "build_info",
					Help: fmt.Sprintf(
						"A metric with a constant '1' value labeled by version, revision, branch, and goversion from which %s was built.",
						c.Application.Name,
					),
					ConstLabels: prometheus.Labels{
						"version":   c.Version,
						"revision":  c.Revision,
						"branch":    c.Branch,
						"goversion": c.GoVersion,
					},
				},
			),
			MetricsTotal: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace: c.Application.Namespace,
					Name:      "metrics_total",
					Help:      "The total number of metrics to be scraped and defined into YAML files.",
				},
			),
			ScrapesSuccess: prometheus.NewCounter(
				prometheus.CounterOpts{
					Namespace:   c.Application.Namespace,
					Name:        "scrapes_success_total",
					Help:        "The total number of times of AWS CloudWatch API scraped for metrics with successful results.",
					ConstLabels: nil,
				},
			),
			ScrapesErrors: prometheus.NewCounter(
				prometheus.CounterOpts{
					Namespace:   c.Application.Namespace,
					Subsystem:   "collector",
					Name:        "scrapes_errors_total",
					Help:        "The total number of times of AWS CloudWatch API scraped for metrics with error results.",
					ConstLabels: nil,
				},
			),
			ScrapesMessages: prometheus.NewCounter(
				prometheus.CounterOpts{
					Namespace:   c.Application.Namespace,
					Subsystem:   "collector",
					Name:        "scrapes_messages_total",
					Help:        "The total number of times of AWS CloudWatch API scraped for metrics and we got some message results. (see exporter logs)",
					ConstLabels: nil,
				},
			),
			MetricsScrapesSuccess: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   c.Application.Namespace,
					Subsystem:   "collector",
					Name:        "metrics_scrapes_success_total",
					Help:        "The total number of metrics of AWS CloudWatch API scraped with successful results.",
					ConstLabels: nil,
				},
			),
			MetricsScrapesErrors: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   c.Application.Namespace,
					Subsystem:   "collector",
					Name:        "metrics_scrapes_errors_total",
					Help:        "The total number of metrics of AWS CloudWatch API scraped with errors results.",
					ConstLabels: nil,
				},
			),
			MetricsScrapesEmpty: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   c.Application.Namespace,
					Subsystem:   "collector",
					Name:        "metrics_scrapes_empty_total",
					Help:        "The total number of metrics of AWS CloudWatch API scraped with empty results.",
					ConstLabels: nil,
				},
			),
			MetricsScrapesMessages: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   c.Application.Namespace,
					Subsystem:   "collector",
					Name:        "metrics_scrapes_messages_total",
					Help:        "The total number of metrics of AWS CloudWatch API scraped and we got some messages results. (see exporter logs)",
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
	ch <- c.ownMetrics.MetricsTotal.Desc()
	c.ownMetrics.ScrapesSuccess.Describe(ch)
	c.ownMetrics.ScrapesErrors.Describe(ch)
	c.ownMetrics.ScrapesMessages.Describe(ch)
	c.ownMetrics.MetricsScrapesSuccess.Describe(ch)
	c.ownMetrics.MetricsScrapesErrors.Describe(ch)
	c.ownMetrics.MetricsScrapesEmpty.Describe(ch)
	c.ownMetrics.MetricsScrapesMessages.Describe(ch)

	// Describe all metrics constructed from yaml files
	for _, md := range c.metrics.GetMetricsDesc() {
		ch <- md
	}
}

// Implements prometheus.Collector Interface
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock() // To protect metrics from concurrent collects.
	defer c.mutex.Unlock()

	// this metrics is constant
	c.ownMetrics.Info.Set(1)
	ch <- c.ownMetrics.Info

	// Going to scrape metrics from yaml files
	c.scrape(ch)
}

// this do the job of scrape the metrics, parse the response from AWS CloudWatch and
// create the prometheus metrics
func (c *Collector) scrape(ch chan<- prometheus.Metric) {
	c.ownMetrics.Up.Set(1)

	// get the timestamps necessary to query metrics from AWS CloudWatch
	//              points     period        now()
	//                ↓        ↓→  ←↓         ↓
	// [(startTime).............................(endTime)] → time
	startTime, endTime, period := metrics.GetTimeStamps(
		time.Now(),
		c.conf.Application.MetricStatPeriod,
		c.conf.Application.MetricTimeWindow)

	// TODO: Implement the paginator using nextoken
	mdi := c.metrics.GetMetricDataInput(startTime, endTime, period, "")

	// number of metrics to be scrape and defined in yaml files
	c.ownMetrics.MetricsTotal.Set(float64(len(mdi.MetricDataQueries)))

	// TODO: Remove it from here and implement a new interface and package to collect metrics
	// but it is here because every time the exporter going to scrape metrics the
	// cloudwatch session is refreshed, I need to be sure that credential never expired
	// when move it from here
	svc := cloudwatch.New(c.sess)

	// Scrape CloudWatch Metrics
	mdo, err := svc.GetMetricData(mdi)
	if err != nil {
		c.ownMetrics.Up.Set(0)
		c.ownMetrics.ScrapesErrors.Inc()
		log.Errorf("Error getting metrics %v", err)
	} else {
		c.ownMetrics.ScrapesSuccess.Inc()
	}

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
			var messages []string
			for _, m := range mdr.Messages {
				messages = append(messages, *m.Value)
			}
			mgsString := strings.Join(messages, ",")
			log.Warnf("Message field for metric id: %s, contain: %s", *mdr.Id, mgsString)
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

		c.ownMetrics.MetricsScrapesSuccess.Inc()

		// Notify metrics to prometheus
		ch <- nm
	}

	// report own metrics
	ch <- c.ownMetrics.Up
	ch <- c.ownMetrics.MetricsTotal
	ch <- c.ownMetrics.ScrapesSuccess
	ch <- c.ownMetrics.ScrapesErrors
	ch <- c.ownMetrics.ScrapesMessages
	ch <- c.ownMetrics.MetricsScrapesSuccess
	ch <- c.ownMetrics.MetricsScrapesErrors
	ch <- c.ownMetrics.MetricsScrapesEmpty
	ch <- c.ownMetrics.MetricsScrapesMessages
}

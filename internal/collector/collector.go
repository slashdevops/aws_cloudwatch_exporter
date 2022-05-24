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
package collector

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/slashdevops/aws_cloudwatch_exporter/internal/config"
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
	svc        *cloudwatch.CloudWatch
	metrics    metrics.Metrics
	mutex      sync.RWMutex
	ownMetrics *OwnMetrics
}

func New(c *config.All, m metrics.Metrics, cwc *cloudwatch.CloudWatch) *Collector {
	return &Collector{
		conf:    c,
		svc:     cwc,
		metrics: m,
		ownMetrics: &OwnMetrics{
			Up: prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace: c.Application.Name,
				Name:      "up",
				Help:      "Was the last scrape of " + c.Application.Name + " successful.",
			}),
			Info: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace: c.Application.Name,
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
					Namespace: c.Application.Name,
					Name:      "metrics_total",
					Help:      "The total number of metrics to be scraped and was defined as metrics queries files.",
				},
			),
			ScrapesSuccess: prometheus.NewCounter(
				prometheus.CounterOpts{
					Namespace:   c.Application.Name,
					Name:        "scrapes_success_total",
					Help:        "The total number of times AWS CloudWatch API scraped for metrics with successful results.",
					ConstLabels: nil,
				},
			),
			ScrapesErrors: prometheus.NewCounter(
				prometheus.CounterOpts{
					Namespace:   c.Application.Name,
					Subsystem:   "collector",
					Name:        "scrapes_errors_total",
					Help:        "The total number of times AWS CloudWatch API scraped for metrics with error results.",
					ConstLabels: nil,
				},
			),
			ScrapesMessages: prometheus.NewCounter(
				prometheus.CounterOpts{
					Namespace:   c.Application.Name,
					Subsystem:   "collector",
					Name:        "scrapes_messages_total",
					Help:        "The total number of times AWS CloudWatch API scraped for metrics and we got some message results. (see exporter logs)",
					ConstLabels: nil,
				},
			),
			MetricsScrapesSuccess: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   c.Application.Name,
					Subsystem:   "collector",
					Name:        "metrics_scrapes_success_total",
					Help:        "The total number of metrics AWS CloudWatch API scraped with successful results.",
					ConstLabels: nil,
				},
			),
			MetricsScrapesErrors: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   c.Application.Name,
					Subsystem:   "collector",
					Name:        "metrics_scrapes_errors_total",
					Help:        "The total number of metrics AWS CloudWatch API scraped with errors results.",
					ConstLabels: nil,
				},
			),
			MetricsScrapesEmpty: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   c.Application.Name,
					Subsystem:   "collector",
					Name:        "metrics_scrapes_empty_total",
					Help:        "The total number of metrics AWS CloudWatch API scraped with empty results.",
					ConstLabels: nil,
				},
			),
			MetricsScrapesMessages: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Namespace:   c.Application.Name,
					Subsystem:   "collector",
					Name:        "metrics_scrapes_messages_total",
					Help:        "The total number of metrics AWS CloudWatch API scraped and we got some messages results. (see exporter logs)",
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

	// Describe all metrics constructed from metrics queries files
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

	// Scrape AWS CloudWatch Metrics
	mdo, err := c.svc.GetMetricData(mdi)
	if err != nil {
		c.ownMetrics.Up.Set(0)
		c.ownMetrics.ScrapesErrors.Inc()
		log.Errorf("Error getting AWS CloudWatch Metrics %v", err)
	} else {
		c.ownMetrics.ScrapesSuccess.Inc()
	}

	// Some information came from the metrics scrape
	// could be and error or a paginator message
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
			log.Errorf("Error gotten when scrap metric id: %s, label: %s. Check your metrics queries files.", *mdr.Id, *mdr.Label)
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
			log.Warnf("Message field for metric id: %s, contain: %s. Check your metrics queries files.", *mdr.Id, mgsString)
		}

		// no metric value came, continue with the next
		if len(mdr.Values) == 0 {
			c.ownMetrics.MetricsScrapesEmpty.Inc()
			log.Warnf("No values gotten for metric id: %s. Check your metrics queries files.", *mdr.Id)
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

		// Notify scraped metrics to prometheus
		ch <- nm
	}

	// Notify own metrics
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

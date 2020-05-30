package metrics

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/slashdevops/aws_cloudwatch_exporter/config"
)

type Metrics interface {
	GetMetricDataInput(time.Time, time.Time, time.Duration, string) *cloudwatch.GetMetricDataInput
}

type metrics struct {
	*config.MetricDataQueriesConf
}

func New(mq *config.MetricDataQueriesConf) Metrics {
	return &metrics{
		mq,
	}
}

// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/using-metric-math.html#metric-math-syntax
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html
func (m *metrics) GetMetricDataInput(st time.Time, et time.Time, p time.Duration, nt string) *cloudwatch.GetMetricDataInput {
	dataQry := m.getMetricDataQuery(p)

	var mdi *cloudwatch.GetMetricDataInput

	if len(nt) > 0 {
		mdi = &cloudwatch.GetMetricDataInput{
			StartTime:         aws.Time(st),
			EndTime:           aws.Time(et),
			MetricDataQueries: dataQry,
			NextToken:         aws.String(nt),
			ScanBy:            aws.String(cloudwatch.ScanByTimestampDescending), // Get the fresh data first
		}
	} else {
		mdi = &cloudwatch.GetMetricDataInput{
			StartTime:         aws.Time(st),
			EndTime:           aws.Time(et),
			MetricDataQueries: dataQry,
			ScanBy:            aws.String(cloudwatch.ScanByTimestampDescending), // Get the fresh data first
		}
	}
	return mdi
}

// This function is used to transform the structure config.MetricDataQueriesConf which contains
// the values read from config file metrics.yaml to a cloudwatch.MetricDataQuery structure which is
// the default structure used to get cloudwatch metrics data
func (m *metrics) getMetricDataQuery(p time.Duration) []*cloudwatch.MetricDataQuery {

	// time.Duration is in nanoseconds, and the CW API need it in seconds
	period := int64(p / time.Second)

	var dataQry []*cloudwatch.MetricDataQuery

	for _, m := range m.MetricDataQueries {

		// Fill the internal struct with dimension
		var dimQry []*cloudwatch.Dimension
		for _, dim := range m.MetricStat.Metric.Dimensions {
			d := &cloudwatch.Dimension{
				Name:  aws.String(dim.Name),
				Value: aws.String(dim.Value),
			}
			dimQry = append(dimQry, d)
		}

		metricsQry := &cloudwatch.MetricDataQuery{
			Id:    aws.String(m.ID),
			Label: aws.String(m.MetricStat.Metric.Namespace + " " + m.MetricStat.Metric.MetricName + " " + m.MetricStat.Stat), // will be used to set prometheus metric name
			MetricStat: &cloudwatch.MetricStat{
				Metric: &cloudwatch.Metric{
					Dimensions: dimQry,
					MetricName: aws.String(m.MetricStat.Metric.MetricName),
					Namespace:  aws.String(m.MetricStat.Metric.Namespace),
				},
				// Period: aws.Int64(m.MetricStat.Period),
				Period: aws.Int64(period),
				Stat:   aws.String(m.MetricStat.Stat),
			},
			// Period:     aws.Int64(period),
			ReturnData: aws.Bool(true), // Return the timestamps and raw data values of this metric.
		}
		dataQry = append(dataQry, metricsQry)
	}
	return dataQry
}

func (m *metrics) GetPrometheusMetrics(mdo *cloudwatch.GetMetricDataOutput) []prometheus.Metric {

	/*
		if len(mdo.MetricDataResults) < 0 {

		}
	*/

	var promMetrics []prometheus.Metric

	for _, mdr := range mdo.MetricDataResults {

		pmn := prometheus.BuildFQName()

		des := prometheus.NewDesc(
			pmn,
			"",
			[]string{""},
			nil)

		for i, val := range mdr.Values {
			pm := prometheus.MustNewConstMetric(des, prometheus.GaugeValue, *val, "")
			pm = prometheus.NewMetricWithTimestamp(*mdr.Timestamps[i], pm)
			promMetrics = append(promMetrics, pm)
		}
	}

	return promMetrics
}

// Return the necessary inputs for function NewGetMetricDataInput
func GetTimeStamps(t time.Time, p string) (startTime time.Time, endTime time.Time, period time.Duration) {
	period, err := time.ParseDuration(p)
	if err != nil {
		log.Errorf("Error parsing period: %v, %v", p, err)
	}

	// endTime = t.Truncate(period)
	// startTime = t.Truncate(period).Add(period * -1)
	endTime = t.Truncate(period).Add(period * 1)
	startTime = t.Truncate(period).Add(period * -2)
	return
}

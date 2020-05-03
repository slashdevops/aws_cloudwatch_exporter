package metrics

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/slashdevops/aws_cloudwatch_exporter/config"
)

// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/using-metric-math.html#metric-math-syntax
func NewMetrics(c *config.MetricsQueriesConf) *cloudwatch.GetMetricDataInput {

	dataQry := getAWSDataQuery(c)

	return &cloudwatch.GetMetricDataInput{
		EndTime:           nil,
		MaxDatapoints:     nil,
		MetricDataQueries: dataQry,
		NextToken:         nil,
		ScanBy:            nil,
		StartTime:         nil,
	}
}

func getAWSDataQuery(c *config.MetricsQueriesConf) []*cloudwatch.MetricDataQuery {

	periodQry := int64(60 * 5)
	var dataQry []*cloudwatch.MetricDataQuery

	for _, m := range c.MetricDataQueries {

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
			Id: aws.String(m.ID),
			MetricStat: &cloudwatch.MetricStat{
				Metric: &cloudwatch.Metric{
					Dimensions: dimQry,
					MetricName: aws.String(m.MetricStat.Metric.MetricName),
					Namespace:  aws.String(m.MetricStat.Metric.Namespace),
				},
				Period: aws.Int64(m.MetricStat.Period),
				Stat:   aws.String(m.MetricStat.Stat),
			},
			Period:     aws.Int64(periodQry),
			ReturnData: aws.Bool(true), //to return the timestamps and raw data values of this metric.
		}
		dataQry = append(dataQry, metricsQry)
	}
	return dataQry
}

func getPrometheusMetrics(mdo *cloudwatch.GetMetricDataOutput) []prometheus.Metric {

	/*if len(mdo.Messages) < 0 {

	}*/
	var promMetrics []prometheus.Metric

	for _, mr := range mdo.MetricDataResults {
		mn := prometheus.BuildFQName()

		des := prometheus.NewDesc(mn, mr.Label)
		m := prometheus.MustNewConstMetric(des)
		prometheus.NewMetricWithTimestamp()
		append(promMetrics, m)
	}

	return promMetrics
}

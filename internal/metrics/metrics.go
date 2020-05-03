package metrics

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/slashdevops/aws_cloudwatch_exporter/config"
)

func NewMetrics(c *config.MetricsQueriesConf) (*cloudwatch.GetMetricDataInput, error) {

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

	metricsDataInputQry := &cloudwatch.GetMetricDataInput{
		EndTime:           nil,
		MaxDatapoints:     nil,
		MetricDataQueries: dataQry,
		NextToken:         nil,
		ScanBy:            nil,
		StartTime:         nil,
	}

	return metricsDataInputQry, nil
}

package metrics

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/slashdevops/aws_cloudwatch_exporter/config"
)

// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/using-metric-math.html#metric-math-syntax
func NewGetMetricDataInput(c *config.MetricsQueriesConf, st time.Time, et time.Time, p int64, nt string) *cloudwatch.GetMetricDataInput {

	dataQry := getAWSDataQuery(c, p)
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

// This function is used to transform the structure config.MetricsQueriesConf whcih contains
// the values read from config file metrics.yaml to a cloudwatch.MetricDataQuery structure which is
// the default structure used to get cloudwatch metrics data
func getAWSDataQuery(c *config.MetricsQueriesConf, p int64) []*cloudwatch.MetricDataQuery {

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
			Period:     aws.Int64(p),
			ReturnData: aws.Bool(true), // Return the timestamps and raw data values of this metric.
		}
		dataQry = append(dataQry, metricsQry)
	}
	return dataQry
}

/*func getPrometheusMetrics(mdo *cloudwatch.GetMetricDataOutput) []prometheus.Metric {

	//if len(mdo.Messages) < 0 {
	//}
	var promMetrics []prometheus.Metric

	for _, mr := range mdo.MetricDataResults {
		mn := prometheus.BuildFQName()

		des := prometheus.NewDesc(mn, mr.Label)
		m := prometheus.MustNewConstMetric(des)
		prometheus.NewMetricWithTimestamp()
		promMetrics = append(promMetrics, m)
	}

	return promMetrics
}
*/

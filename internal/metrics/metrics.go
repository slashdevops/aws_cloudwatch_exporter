package metrics

import (
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/slashdevops/aws_cloudwatch_exporter/config"
	"github.com/slashdevops/aws_cloudwatch_exporter/internal/camelcase"
	"golang.org/x/tools/go/ssa/interp/testdata/src/fmt"
)

type Metrics interface {
	// Used to assemble the AWS GetMetricDataInput data structure
	GetMetricDataInput(time.Time, time.Time, time.Duration, string) *cloudwatch.GetMetricDataInput

	//
	SetMetric(id string, metric prometheus.Metric)
	GetMetricDesc(id string) *prometheus.Desc
	GetMetricsDesc() map[string]*prometheus.Desc
}

type metrics struct {
	// Metrics queries structure assemble from metrics queries yaml files
	MetricDataQueriesConf *config.MetricDataQueriesConf

	// The prometheus metrics created from MetricDataQueriesConf but without values
	// The prometheus metrics created from MetricDataQueriesConf but without values
	PrometheusMetricsDesc map[string]*prometheus.Desc
	PrometheusMetrics     map[string]prometheus.Metric
}

func New(conf *config.All) Metrics {
	return &metrics{
		MetricDataQueriesConf: &conf.MetricDataQueriesConf,
		PrometheusMetricsDesc: createPrometheusMetricsDesc(conf),
		PrometheusMetrics:     make(map[string]prometheus.Metric),
	}
}

// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_concepts.html
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

	for _, m := range m.MetricDataQueriesConf.MetricDataQueries {

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
			Label: aws.String(camelcase.ToSnake(m.MetricStat.Metric.Namespace) + "_" + camelcase.ToSnake(m.MetricStat.Metric.MetricName) + "_" + camelcase.ToSnake(m.MetricStat.Stat)),
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

func (m *metrics) SetMetric(id string, metric prometheus.Metric) {
	m.PrometheusMetrics[id] = metric
}

func (m *metrics) GetMetricDesc(id string) *prometheus.Desc {
	i, ok := m.PrometheusMetricsDesc[id]
	if !ok {
		log.Errorf("Metric id: %s does not exist", id)
	}
	return i
}

func (m *metrics) GetMetricsDesc() map[string]*prometheus.Desc {
	return m.PrometheusMetricsDesc
}

// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_concepts.html
// Create the prometheus metrics scaffolding without any value
//   - Id: m1
//    MetricStat:
//      Metric:
//        Namespace: AWS/EC2
//        MetricName: CPUUtilization            --> AWS/EC2_CPUUtilization_Average{job="aws_cloudwatch_exporter", instance="", AutoScalingGroupName="eks-prod-01-apps-01-asg"} value_from_scrap
//        Dimensions:                               aws_ec2_cpu_utilization_average{job="aws_cloudwatch_exporter", instance="", auto_scaling_group_name="eks-prod-01-apps-01-asg"} value_from_scrap
//          - Name: AutoScalingGroupName            aws_cloudwatch_exporter_aws_ec2_cpu_utilization_average{job="aws_cloudwatch_exporter", instance="", auto_scaling_group_name="eks-prod-01-apps-01-asg"} value_from_scrap
//            Value: eks-prod-01-apps-01-asg
//      Stat: Average
func createPrometheusMetricsDesc(conf *config.All) map[string]*prometheus.Desc {
	mdqc := conf.MetricDataQueriesConf
	promMetricsDesc := make(map[string]*prometheus.Desc)

	var helpTmpl string = "%s represent the AWS CloudWatch Metric: %s --> %s, Dimensions: [%s], Statistic: %s"

	// for every metric query defined into the yaml files
	for _, mdq := range mdqc.MetricDataQueries {

		// Add dimensions as prometheus metric labels
		mcl := make(prometheus.Labels)
		for _, v := range mdq.MetricStat.Metric.Dimensions {
			mcl[v.Name] = v.Value
		}
		var dimKeys []string
		for k := range mcl {
			dimKeys = append(dimKeys, k)
		}

		mn := camelcase.ToSnake(mdq.MetricStat.Metric.Namespace) + "_" + camelcase.ToSnake(mdq.MetricStat.Metric.MetricName) + "_" + camelcase.ToSnake(mdq.MetricStat.Stat)
		hs := fmt.Sprint(helpTmpl, mn, mdq.MetricStat.Metric.Namespace, mdq.MetricStat.Metric.MetricName, strings.Join(dimKeys, ","), mdq.MetricStat.Stat)

		promMetricsDesc[mdq.ID] = prometheus.NewDesc(mn, hs, nil, mcl)
	}

	return promMetricsDesc
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
